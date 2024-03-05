package connpool

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
	"webook/pkg/logger"
)

type DoubleWritePool struct {
	src gorm.ConnPool
	dst gorm.ConnPool

	pattern *atomicx.Value[string]

	l logger.LoggerV1
}

func NewDoubleWritePool(src *gorm.DB, dst *gorm.DB, l logger.LoggerV1) *DoubleWritePool {
	return &DoubleWritePool{src: src.ConnPool,
		dst:     dst.ConnPool,
		l:       l,
		pattern: atomicx.NewValueOf(PatternSrcOnly),
	}
}

func (d *DoubleWritePool) UpdatePattern(pattern string) error {
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst, PatternDstFirst, PatternDstOnly:
		d.pattern.Store(pattern)
		return nil
	default:
		return errUnKnownPattern
	}
}

func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{src: src, l: d.l, pattern: pattern}, err
	case PatternSrcFirst:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			// 同样的，只需要源库事务开成功即可，这里只记录日志
			// 当然，此时也可以考虑回滚掉 src 并返回 error
			d.l.Error("双写阶段，开启目标表事务失败", logger.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, l: d.l, pattern: pattern}, nil
	case PatternDstFirst:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			// 同样的，只需要源库事务开成功即可，这里只记录日志
			// 当然，此时也可以考虑回滚掉 dst 并返回 error
			d.l.Error("双写阶段，开启源表事务失败", logger.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, l: d.l, pattern: pattern}, nil
	case PatternDstOnly:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{dst: dst, l: d.l, pattern: pattern}, err
	default:
		return nil, errUnKnownPattern
	}
}

func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic("双写模式不支持")
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil {
			_, er := d.dst.ExecContext(ctx, query, args...)
			if er != nil {
				d.l.Error("SrcFirst阶段，写入 dst 失败", logger.Error(er), logger.String("sql", query))
			}
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil {
			_, er := d.src.ExecContext(ctx, query, args...)
			if er != nil {
				d.l.Error("DstFirst阶段，写入 src 失败", logger.Error(er), logger.String("sql", query))
			}
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, errUnKnownPattern
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		res, err := d.src.QueryContext(ctx, query, args...)
		if err == nil {
			go func() {
				resDst, er := d.dst.QueryContext(ctx, query, args...)
				if er == nil {
					if res != resDst {
						// 记录日志并通知修复程序
					}
				}
			}()
		}
		return res, err
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnKnownPattern
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		return &sql.Row{}
		// 由于没有带上错误信息，sql.Row 的 err 是私有的，因此可以考虑
		// 使用 unsafe
		// 直接 panic 掉：panic(errUnknownPattern)，但一定要保证上层业务中有处理这个 panic 的逻辑，即 recover()
	}
}

type DoubleWriteTx struct {
	src *sql.Tx
	dst *sql.Tx

	// 事务一定是稳定不变的
	pattern string

	l logger.LoggerV1
}

func (d *DoubleWriteTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternSrcFirst:
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst != nil {
			er := d.dst.Commit()
			if er != nil {
				d.l.Error("SrcFirst阶段，目标表提交事务失败")
			}
		}
		return nil
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src != nil {
			er := d.src.Commit()
			if er != nil {
				d.l.Error("DstFirst阶段，源表提交事务失败")
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Commit()
	default:
		return errUnKnownPattern
	}
}

func (d *DoubleWriteTx) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst != nil {
			er := d.dst.Rollback()
			if er != nil {
				d.l.Error("SrcFirst阶段，目标表提交事务失败")
			}
		}
		return nil
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			er := d.src.Rollback()
			if er != nil {
				d.l.Error("DstFirst阶段，源表提交事务失败")
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Rollback()
	default:
		return errUnKnownPattern
	}
}

func (d *DoubleWriteTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic("双写模式不支持")
}

func (d *DoubleWriteTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		// d.dst != nil: 防止开启事务阶段 (BeginTx)，一个开启另一个失败的情况
		if err == nil && d.dst != nil {
			_, er := d.dst.ExecContext(ctx, query, args...)
			if er != nil {
				d.l.Error("SrcFirst阶段，写入 dst 失败", logger.Error(er), logger.String("sql", query))
			}
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		// d.src != nil: 防止开启事务阶段 (BeginTx)，一个开启另一个失败的情况
		if err == nil && d.src != nil {
			_, er := d.src.ExecContext(ctx, query, args...)
			if er != nil {
				d.l.Error("DstFirst阶段，写入 src 失败", logger.Error(er), logger.String("sql", query))
			}
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, errUnKnownPattern
	}
}

func (d *DoubleWriteTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		res, err := d.src.QueryContext(ctx, query, args...)
		if err == nil {
			go func() {
				resDst, er := d.dst.QueryContext(ctx, query, args...)
				if er == nil {
					if res != resDst {
						// 记录日志并通知修复程序
					}
				}
			}()
		}
		return res, err
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnKnownPattern
	}
}

func (d *DoubleWriteTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		return &sql.Row{}
		// 由于没有带上错误信息，sql.Row 的 err 是私有的，因此可以考虑
		// 使用 unsafe
		// 直接 panic 掉：panic(errUnknownPattern)，但一定要保证上层业务中有处理这个 panic 的逻辑，即 recover()
	}
}

var errUnKnownPattern = errors.New("未知双写模式 pattern ")

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)

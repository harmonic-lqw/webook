package dao

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
	"webook/pkg/logger"
)

// DoubleWriteDAO 直接修改 DAO 的双写的实现，更加适用于异构、个性化、业务强耦合的场景
type DoubleWriteDAO struct {
	src InteractiveDAO
	dst InteractiveDAO

	// 用原子的思路，因为后面这个是动态更新的，需要保证并发安全
	pattern *atomicx.Value[string]

	l logger.LoggerV1
}

func (d *DoubleWriteDAO) GetTopNLike(ctx context.Context) ([]Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func NewDoubleWriteDAO(src *gorm.DB, dst *gorm.DB, l logger.LoggerV1) InteractiveDAO {
	return &DoubleWriteDAO{src: NewGORMInteractiveDAO(src),
		dst:     NewGORMInteractiveDAO(dst),
		l:       l,
		pattern: atomicx.NewValueOf(PatternSrcOnly),
	}
}

func (d *DoubleWriteDAO) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWriteDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, bizId)
	case PatternSrcFirst:
		err := d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// 这里最好不要 return err
			// 因为在这个阶段，src 成功在业务上就算成功了
			d.l.Error("SrcFirst阶段，写入 dst 失败", logger.Error(err),
				logger.Int64("biz_id", bizId),
				logger.String("biz", biz))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, bizId)
		if err == nil {
			er := d.src.IncrReadCnt(ctx, biz, bizId)
			if er != nil {
				d.l.Error("DstFirst阶段，写入 src 失败", logger.Error(er),
					logger.Int64("biz_id", bizId),
					logger.String("biz", biz))
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, bizId)
	default:
		return errUnKnownPattern
	}
}

func (d *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.BatchIncrReadCnt(ctx, bizs, bizIds)
	case PatternSrcFirst:
		err := d.src.BatchIncrReadCnt(ctx, bizs, bizIds)
		if err != nil {
			return err
		}
		err = d.dst.BatchIncrReadCnt(ctx, bizs, bizIds)
		if err != nil {
			d.l.Error("SrcFirst阶段，写入 dst 失败", logger.Error(err))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.BatchIncrReadCnt(ctx, bizs, bizIds)
		if err == nil {
			er := d.src.BatchIncrReadCnt(ctx, bizs, bizIds)
			if er != nil {
				d.l.Error("DstFirst阶段，写入 src 失败", logger.Error(er))
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.BatchIncrReadCnt(ctx, bizs, bizIds)
	default:
		return errUnKnownPattern
	}
}

func (d *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.InsertLikeInfo(ctx, biz, bizId, uid)
	case PatternSrcFirst:
		err := d.src.InsertLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			return err
		}
		err = d.dst.InsertLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("SrcFirst阶段，写入 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.InsertLikeInfo(ctx, biz, bizId, uid)
		if err == nil {
			er := d.src.InsertLikeInfo(ctx, biz, bizId, uid)
			if er != nil {
				d.l.Error("DstFirst阶段，写入 src 失败", logger.Error(er),
					logger.String("biz", biz),
					logger.Int64("biz_id", bizId),
					logger.Int64("uid", uid))
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.InsertLikeInfo(ctx, biz, bizId, uid)
	default:
		return errUnKnownPattern
	}
}

func (d *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.DeleteLikeInfo(ctx, biz, bizId, uid)
	case PatternSrcFirst:
		err := d.src.DeleteLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			return err
		}
		err = d.dst.DeleteLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("SrcFirst阶段，写入 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.DeleteLikeInfo(ctx, biz, bizId, uid)
		if err == nil {
			er := d.src.DeleteLikeInfo(ctx, biz, bizId, uid)
			if er != nil {
				d.l.Error("DstFirst阶段，写入 src 失败", logger.Error(er),
					logger.String("biz", biz),
					logger.Int64("biz_id", bizId),
					logger.Int64("uid", uid))
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.DeleteLikeInfo(ctx, biz, bizId, uid)
	default:
		return errUnKnownPattern
	}
}

func (d *DoubleWriteDAO) InsertCollectionBiz(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.InsertCollectionBiz(ctx, biz, cid, bizId, uid)
	case PatternSrcFirst:
		err := d.src.InsertCollectionBiz(ctx, biz, cid, bizId, uid)
		if err != nil {
			return err
		}
		err = d.dst.InsertCollectionBiz(ctx, biz, cid, bizId, uid)
		if err != nil {
			d.l.Error("SrcFirst阶段，写入 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("cid", cid),
				logger.Int64("uid", uid))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.InsertCollectionBiz(ctx, biz, cid, bizId, uid)
		if err == nil {
			er := d.src.InsertCollectionBiz(ctx, biz, cid, bizId, uid)
			if er != nil {
				d.l.Error("DstFirst阶段，写入 src 失败", logger.Error(er),
					logger.String("biz", biz),
					logger.Int64("biz_id", bizId),
					logger.Int64("cid", cid),
					logger.Int64("uid", uid))
			}
		}
		return err
	case PatternDstOnly:
		return d.dst.InsertCollectionBiz(ctx, biz, cid, bizId, uid)
	default:
		return errUnKnownPattern
	}
}

func (d *DoubleWriteDAO) GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		uLikeBiz, err := d.src.GetLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("SrcOnly阶段，读 src 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
			return UserLikeBiz{}, nil
		}
		return uLikeBiz, err
	case PatternSrcFirst:
		uLikeBiz, err := d.src.GetLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("SrcFirst阶段，读 src 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
			return UserLikeBiz{}, nil
		}
		return uLikeBiz, err
	case PatternDstFirst:
		uLikeBiz, err := d.dst.GetLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("DstFirst阶段，读 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
			return UserLikeBiz{}, nil
		}
		return uLikeBiz, err
	case PatternDstOnly:
		uLikeBiz, err := d.dst.GetLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("DstOnly阶段，读 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
			return UserLikeBiz{}, nil
		}
		return uLikeBiz, err
	default:
		return UserLikeBiz{}, errUnKnownPattern
	}
}

func (d *DoubleWriteDAO) GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		uColBiz, err := d.src.GetCollectInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("SrcOnly阶段，读 src 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
			return UserCollectionBiz{}, nil
		}
		return uColBiz, err
	case PatternSrcFirst:
		uColBiz, err := d.src.GetCollectInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("SrcFirst阶段，读 src 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
			return UserCollectionBiz{}, nil
		}
		return uColBiz, err
	case PatternDstFirst:
		uColBiz, err := d.dst.GetCollectInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("DstFirst阶段，读 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
			return UserCollectionBiz{}, nil
		}
		return uColBiz, err
	case PatternDstOnly:
		uColBiz, err := d.dst.GetCollectInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("DstOnly阶段，读 dst 失败", logger.Error(err),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
				logger.Int64("uid", uid))
			return UserCollectionBiz{}, nil
		}
		return uColBiz, err
	default:
		return UserCollectionBiz{}, errUnKnownPattern
	}
}

func (d *DoubleWriteDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.Get(ctx, biz, bizId)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.Get(ctx, biz, bizId)
	default:
		return Interactive{}, errUnKnownPattern
	}
}

func (d *DoubleWriteDAO) GetByIds(ctx context.Context, biz string, bizIds []int64) ([]Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.GetByIds(ctx, biz, bizIds)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.GetByIds(ctx, biz, bizIds)
	default:
		return []Interactive{}, errUnKnownPattern
	}
}

var errUnKnownPattern = errors.New("未知双写模式 pattern ")

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)

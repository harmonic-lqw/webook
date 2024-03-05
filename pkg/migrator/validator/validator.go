package validator

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

type Validator[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	l         logger.LoggerV1
	producer  events.Producer
	direction string

	batchSize int
	// assignment 12
	batchSizeBase int

	// 使用 utime 和 sleepInterval 的组合 就可以实现全量校验和增量校验的控制
	utime         int64
	sleepInterval time.Duration

	// 借助该字段抽取全量校验和增量校验的逻辑
	fromBase func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, direction string, l logger.LoggerV1, producer events.Producer) *Validator[T] {
	return &Validator[T]{
		base:      base,
		target:    target,
		l:         l,
		producer:  producer,
		direction: direction}
}

func (v *Validator[T]) Utime(utime int64) *Validator[T] {
	v.utime = utime
	return v
}

func (v *Validator[T]) SleepInterval(sleepInterval time.Duration) *Validator[T] {
	v.sleepInterval = sleepInterval
	return v
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	//err := v.validateBaseToTarget(ctx)
	//if err != nil {
	//	return err
	//}
	//return v.validateTargetToBase(ctx)

	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

// week13 assignment 12
func (v *Validator[T]) validateBaseToTargetBatch(ctx context.Context) error {
	offset := 0
	for {
		var bases []T
		err := v.base.WithContext(ctx).
			Select("id").
			Order("id").
			Offset(offset).
			Limit(v.batchSizeBase).
			Find(&bases).Error

		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound || len(bases) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			// 查询出错
			v.l.Error("base -> target 批量查询 base 失败", logger.Error(err))
			offset += len(bases)
			continue
		}

		ids := slice.Map(bases, func(idx int, src T) int64 {
			return src.ID()
		})
		var tars []T
		err = v.target.WithContext(ctx).Select("id").
			Where("id IN ?", ids).
			Find(&tars).Error

		if err == gorm.ErrRecordNotFound || len(tars) == 0 {
			v.notifyTargetMissing(bases)
			offset += len(bases)
			continue
		}
		if err != nil {
			v.l.Error("base -> target 批量查询 target 失败", logger.Error(err))
			offset += len(bases)
			continue
		}
		diff := slice.DiffSetFunc(bases, tars, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyTargetMissing(diff)
		if len(bases) < v.batchSizeBase {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		offset += len(bases)

	}
}

func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := 0
	for {
		// 这里可以接入一些负载检测，当负载低的时候才进行数据校验

		src, err := v.fromBase(ctx, offset)
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound {
			// 没有数据了
			if v.sleepInterval <= 0 {
				// 全量校验
				return nil
			}
			// 增量校验
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			// 查询出错
			v.l.Error("base -> target 查询 base 失败", logger.Error(err))
			offset += 1
			continue
		}

		var dst T
		err = v.target.WithContext(ctx).
			Where("id = ?", src.ID()).
			First(&dst).
			Error

		switch err {
		case gorm.ErrRecordNotFound:
			// 需要同步，丢一条消息到 kafka
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		case nil:
			equal := src.CompareTo(dst)
			if !equal {
				// 需要同步，丢一条消息到 kafka
				v.notify(src.ID(), events.InconsistentEventTypeNotEqual)
			}
		default:
			v.l.Error("base -> target 查询 target 失败",
				logger.Int64("id", src.ID()),
				logger.Error(err))
		}
		offset += 1
	}
}

func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := 0
	for {

		var tars []T
		// target2base 的增量校验，其实也可以直接用 id 来索引，反正很快
		err := v.target.WithContext(ctx).
			Select("id").
			Order("id").
			Offset(offset).
			Limit(v.batchSize).
			Find(&tars).
			Error
		//err := v.target.WithContext(ctx).
		//	Select("id").
		//	Where("utime > ?", v.utime).
		//	Order("utime").
		//	Offset(offset).
		//	Limit(v.batchSize).
		//	Find(&tars).
		//	Error
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound || len(tars) == 0 {
			// 检索完 target 没数据了
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			v.l.Error("target -> base 查询 target 失败", logger.Error(err))
			offset += len(tars)
			continue
		}

		// 开始检索 base
		var bases []T
		ids := slice.Map(tars, func(idx int, t T) int64 {
			return t.ID()
		})
		err = v.base.WithContext(ctx).Select("id").
			Where("id IN ?", ids).
			Find(&bases).Error

		if len(bases) == 0 {
			v.notifyBaseMissing(tars)
			offset += len(tars)
			continue
		}
		switch err {
		case gorm.ErrRecordNotFound:
			v.notifyBaseMissing(tars)
		case nil:
			// 求差集，diff 中即 target 有但 base 没有
			diff := slice.DiffSetFunc(tars, bases, func(src, dst T) bool {
				return src.ID() == dst.ID()
			})
			v.notifyBaseMissing(diff)
		default:
			v.l.Error("target -> base 查询 base 失败", logger.Error(err))
		}

		// 批次不够，发送完消息后退出
		if len(tars) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		offset += len(tars)
	}
}

func (v *Validator[T]) Full() *Validator[T] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).
		Order("id").
		Offset(offset).
		First(&src).
		Error
	return src, err
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).
		Where("utime > ?", v.utime).
		Order("utime").
		Offset(offset).
		First(&src).
		Error
	return src, err
}

func (v *Validator[T]) notifyBaseMissing(tars []T) {
	for _, val := range tars {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notifyTargetMissing(bases []T) {
	for _, val := range bases {
		v.notify(val.ID(), events.InconsistentEventTypeTargetMissing)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Type:      typ,
		Direction: v.direction,
	})
	v.l.Error("发送不一致消息到 kafka 失败",
		logger.Error(err),
		logger.String("type", typ),
		logger.Int64("id", id))
}

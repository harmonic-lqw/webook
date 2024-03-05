package fixer

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

type Fixer[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	columns []string
}

func NewFixerV1[T migrator.Entity](base *gorm.DB, target *gorm.DB, columns []string) *Fixer[T] {
	return &Fixer[T]{base: base,
		target:  target,
		columns: columns,
	}
}

func NewFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*Fixer[T], error) {
	rows, err := base.Model(new(T)).Order("id").Rows()
	if err != nil {
		return nil, err

	}
	columns, err := rows.Columns()
	return &Fixer[T]{base: base,
		target:  target,
		columns: columns,
	}, nil
}

// Fix 最简单粗暴的写法，直接去 base 里面找，找到就 upsert，找不到就 delete
func (f *Fixer[T]) Fix(ctx context.Context, id int64) error {
	var t T
	err := f.base.WithContext(ctx).Where("id=?", id).First(&t).Error
	switch err {
	case gorm.ErrRecordNotFound:
		return f.target.WithContext(ctx).Model(&t).Delete("id = ?", id).Error
	case nil:
		// upsert
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	default:
		return err
	}
}

// FixV1 使用 upsert 控制并发安全
func (f *Fixer[T]) FixV1(evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeNotEqual, events.InconsistentEventTypeTargetMissing:
		var t T
		err := f.base.Where("id=?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			return f.target.Model(&t).Delete("id = ?", evt.ID).Error
		case nil:
			// upsert
			return f.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		return f.target.Model(new(T)).Delete("id = ?", evt.ID).Error
	}
	return nil
}

// FixV2 没有并发安全的写法
func (f *Fixer[T]) FixV2(evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeNotEqual:
		var t T
		err := f.base.Where("id=?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			return f.target.Model(&t).Delete("id = ?", evt.ID).Error
		case nil:
			return f.target.Updates(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeTargetMissing:
		var t T
		err := f.base.Where("id=?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			return nil
		case nil:
			// 双写阶段会有并发问题
			return f.target.Create(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		return f.target.Model(new(T)).Delete("id = ?", evt.ID).Error
	}
	return nil
}

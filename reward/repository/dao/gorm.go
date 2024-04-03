package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type RewardGORMDAO struct {
	db *gorm.DB
}

func (dao *RewardGORMDAO) UpdateStatus(ctx context.Context, rid int64, status uint8) error {
	return dao.db.WithContext(ctx).
		Where("id = ?", rid).
		Updates(map[string]any{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func NewRewardGORMDAO(db *gorm.DB) RewardDAO {
	return &RewardGORMDAO{
		db: db,
	}
}

func (dao *RewardGORMDAO) GetReward(ctx context.Context, rid int64) (Reward, error) {
	var r Reward
	err := dao.db.WithContext(ctx).
		Where("id = ?", rid).
		First(&r).Error
	return r, err
}

func (dao *RewardGORMDAO) Insert(ctx context.Context, r Reward) (int64, error) {
	now := time.Now().UnixMilli()
	r.Ctime = now
	r.Utime = now
	err := dao.db.WithContext(ctx).Create(&r).Error
	return r.Id, err
}

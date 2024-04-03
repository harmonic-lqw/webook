package dao

import "context"

type RewardDAO interface {
	Insert(ctx context.Context, r Reward) (int64, error)
	GetReward(ctx context.Context, rid int64) (Reward, error)
	UpdateStatus(ctx context.Context, rid int64, status uint8) error
}

type Reward struct {
	Id      int64  `gorm:"primaryKey, autoIncrement" bson:"id, omitempty"`
	Biz     string `gorm:"index:biz_biz_id"`
	BizId   int64  `gorm:"index:biz_biz_id"`
	BizName string

	TarUid int64 `gorm:"index"`

	Status uint8

	SrcUid int64
	Amount int64
	Ctime  int64
	Utime  int64
}

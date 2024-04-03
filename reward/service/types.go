package service

import (
	"context"
	"webook/reward/domain"
)

type RewardService interface {
	PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error)
	UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error
}

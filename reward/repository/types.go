package repository

import (
	"context"
	"webook/reward/domain"
)

type RewardRepository interface {
	GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	CreateReward(ctx context.Context, r domain.Reward) (int64, error)
	CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error
	GetReward(ctx context.Context, rid int64) (domain.Reward, error)
	UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error
}

package repository

import (
	"context"
	"webook/reward/domain"
	"webook/reward/repository/cache"
	"webook/reward/repository/dao"
)

type rewardRepository struct {
	dao   dao.RewardDAO
	cache cache.RewardCache
}

func (repo *rewardRepository) UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error {
	return repo.dao.UpdateStatus(ctx, rid, status.AsUint8())
}

func (repo *rewardRepository) GetReward(ctx context.Context, rid int64) (domain.Reward, error) {
	r, err := repo.dao.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	return repo.toDomain(r), nil
}

func (repo *rewardRepository) CreateReward(ctx context.Context, r domain.Reward) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(r))
}

func (repo *rewardRepository) GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	return repo.cache.GetCachedCodeURL(ctx, r)
}

func (repo *rewardRepository) CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error {
	return repo.cache.CachedCodeURL(ctx, cu, r)
}

func (repo *rewardRepository) toEntity(r domain.Reward) dao.Reward {
	return dao.Reward{
		Status:  r.Status.AsUint8(),
		Biz:     r.Target.Biz,
		BizId:   r.Target.BizId,
		BizName: r.Target.BizName,
		TarUid:  r.Target.TarUId,
		SrcUid:  r.SrcUid,
		Amount:  r.Amt,
	}
}

func (repo *rewardRepository) toDomain(r dao.Reward) domain.Reward {
	return domain.Reward{
		Id:     r.Id,
		SrcUid: r.SrcUid,
		Target: domain.Target{
			Biz:     r.Biz,
			BizId:   r.BizId,
			BizName: r.BizName,
			TarUId:  r.TarUid,
		},
		Amt:    r.Amount,
		Status: domain.RewardStatus(r.Status),
	}
}

func NewRewardRepository(dao dao.RewardDAO, c cache.RewardCache) RewardRepository {
	return &rewardRepository{dao: dao, cache: c}
}

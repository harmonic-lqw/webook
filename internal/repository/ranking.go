package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
	SetLoad(ctx context.Context, nodeId int64, load int) error
	GetMinLoadNode(ctx context.Context) (int64, int, error)
}

type CachedOnlyRankingRepository struct {
	cache cache.RankingCache

	// V1 结合本地缓存
	redisCache *cache.RankingRedisCache
	localCache *cache.RankingLocalCache
}

func NewCachedOnlyRankingRepositoryV1(redisCache *cache.RankingRedisCache, localCache *cache.RankingLocalCache) RankingRepository {
	return &CachedOnlyRankingRepository{redisCache: redisCache, localCache: localCache}
}

func NewCachedOnlyRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedOnlyRankingRepository{cache: cache}
}

func (r *CachedOnlyRankingRepository) GetMinLoadNode(ctx context.Context) (int64, int, error) {
	return r.cache.GetMinLoadNode(ctx)
}

func (r *CachedOnlyRankingRepository) SetLoad(ctx context.Context, nodeId int64, load int) error {
	return r.cache.SetLoad(ctx, nodeId, load)
}

func (r *CachedOnlyRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return r.cache.Get(ctx)
}

func (r *CachedOnlyRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return r.cache.Set(ctx, arts)
}

// GetTopNV1 结合了本地缓存
func (r *CachedOnlyRankingRepository) GetTopNV1(ctx context.Context) ([]domain.Article, error) {
	res, err := r.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}
	res, err = r.redisCache.Get(ctx)
	if err != nil {
		// 一种提高可用性的方式：给本地缓存兜底，如果 redis 还是没拿到数据，就还是从本地缓存中加载，但此时不需要过期时间这个条件
		return r.localCache.ForceGet(ctx)
	}
	// 回写本地缓存
	_ = r.localCache.Set(ctx, res)
	return res, nil
}

func (r *CachedOnlyRankingRepository) ReplaceTopNV1(ctx context.Context, arts []domain.Article) error {
	_ = r.localCache.Set(ctx, arts)
	return r.cache.Set(ctx, arts)
}

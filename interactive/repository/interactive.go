package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"time"
	"webook/interactive/domain"
	"webook/interactive/repository/cache"
	"webook/interactive/repository/dao"
	"webook/pkg/logger"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	// BatchIncrReadCnt bizs 和 bizIds 长度必须一致
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
	IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) ([]domain.Interactive, error)

	// GetTopNLike assignment week9
	GetTopNLike(ctx context.Context) ([]domain.Interactive, error)
	// SetTopNLike assignment week9
	SetTopNLike(ctx context.Context) error
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.LoggerV1

	// assignment week9
	redisCache *cache.InteractiveRedisCache
	localCache *cache.InteractiveLocalCache
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, bizIds []int64) ([]domain.Interactive, error) {
	intrs, err := c.dao.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Interactive, domain.Interactive](intrs, func(idx int, src dao.Interactive) domain.Interactive {
		return c.toDomain(src)
	}), nil
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	intr, err := c.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, nil
	}

	ie, err := c.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}

	res := c.toDomain(ie)
	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		er := c.cache.Set(newCtx, biz, bizId, res)
		if er != nil {
			c.l.Error("回写缓存失败",
				logger.String("biz", biz),
				logger.Int64("bizId", bizId),
				logger.Error(err))
		}
	}()

	return res, nil
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, biz, bizId, cid, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, bizId)

}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache, l logger.LoggerV1) InteractiveRepository {
	return &CachedInteractiveRepository{dao: dao, cache: cache, l: l}
}

func (c *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	// 更新数据库
	err := c.dao.BatchIncrReadCnt(ctx, bizs, bizIds)
	if err != nil {
		return err
	}
	go func() {
		for i := 0; i < len(bizs); i++ {
			er := c.cache.IncrReadCntIfPresent(ctx, bizs[i], bizIds[i])
			if er != nil {
				// 记录日志
			}
		}
	}()
	return nil
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	// 更新数据库
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	// 更新缓存
	// 可能会因为更新失败导致数据不一致问题，但是问题不大
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) toDomain(ie dao.Interactive) domain.Interactive {
	return domain.Interactive{
		BizId:      ie.BizId,
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
	}
}

// NewCachedInteractiveRepositoryV1 assignment week9
func NewCachedInteractiveRepositoryV1(dao dao.InteractiveDAO, redisCache *cache.InteractiveRedisCache, localCache *cache.InteractiveLocalCache, l logger.LoggerV1) InteractiveRepository {
	return &CachedInteractiveRepository{dao: dao, redisCache: redisCache, localCache: localCache, l: l}
}

// GetTopNLike assignment week9
func (c *CachedInteractiveRepository) GetTopNLike(ctx context.Context) ([]domain.Interactive, error) {
	// 先从本地缓存拿
	res, err := c.localCache.GetTopNLike(ctx)
	if err == nil {
		return res, nil
	}
	// 再从 redis 中拿
	res, err = c.redisCache.GetTopNLike(ctx)
	if err == nil {
		_ = c.localCache.SetTopNLike(ctx, res)
		return res, nil
	}

	// finally get from db
	intrs, err := c.dao.GetTopNLike(ctx)
	if err != nil {
		return []domain.Interactive, err
	}
	var doIntrs []domain.Interactive
	for _, intr := range intrs {
		doIntrs = append(doIntrs, c.toDomain(intr))
	}

	_ = c.SetTopNLike(ctx)
	return doIntrs, nil
}

// SetTopNLike assignment week9
func (c *CachedInteractiveRepository) SetTopNLike(ctx context.Context) error {
	intrs, err := c.dao.GetTopNLike(ctx)
	if err != nil {
		return err
	}
	var doIntrs []domain.Interactive
	for _, intr := range intrs {
		doIntrs = append(doIntrs, c.toDomain(intr))
	}
	_ = c.localCache.SetTopNLike(ctx, doIntrs)
	return c.redisCache.SetTopNLike(ctx, doIntrs)
}

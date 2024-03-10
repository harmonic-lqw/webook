package cache

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"time"
	"webook/interactive/domain"
)

type InteractiveLocalCache struct {
	topLike    *atomicx.Value[[]domain.Article]
	expiration time.Duration
}

func (i *InteractiveLocalCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	//TODO implement me
	panic("implement me")
}

func (i *InteractiveLocalCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	//TODO implement me
	panic("implement me")
}

func (i *InteractiveLocalCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	//TODO implement me
	panic("implement me")
}

func (i *InteractiveLocalCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	//TODO implement me
	panic("implement me")
}

func (i *InteractiveLocalCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func (i *InteractiveLocalCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	//TODO implement me
	panic("implement me")
}

// GetTopNLike assignment week9
func (i *InteractiveLocalCache) GetTopNLike(ctx context.Context) ([]domain.Article, error) {
	arts := i.topLike.Load()
	if len(arts) == 0 {
		return []domain.Article, errors.New("本地未缓存数据")
	}
	return arts, nil
}

// SetTopNLike assignment week9
func (i *InteractiveLocalCache) SetTopNLike(ctx context.Context, arts []domain.Article) error {
	i.topLike.Store(arts)
	return nil
}

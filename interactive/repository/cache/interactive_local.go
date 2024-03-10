package cache

import (
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"time"
	"webook/interactive/domain"
)

type InteractiveLocalCache struct {
	topLike    *atomicx.Value[[]domain.Interactive]
	ddl *atomicx.Value[time.Time]
	expiration time.Duration
}

func NewInteractiveLocalCache(topLike *atomicx.Value[[]domain.Interactive], ddl *atomicx.Value[time.Time]) *InteractiveLocalCache {
	return &InteractiveLocalCache{topLike: topLike,
		ddl: atomicx.NewValue[time.Time](),
		expiration: time.Hour * 24 * 7}
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
func (i *InteractiveLocalCache) GetTopNLike(ctx context.Context) ([]domain.Interactive, error) {
	intrs := i.topLike.Load()
	ddl := i.ddl.Load()
	if len(intrs) == 0 || ddl.Before(time.Now()){
		return []domain.Interactive, errors.New("本地未缓存数据")
	}
	return intrs, nil
}

// SetTopNLike assignment week9
func (i *InteractiveLocalCache) SetTopNLike(ctx context.Context, intrs []domain.Interactive) error {
	i.topLike.Store(intrs)
	i.ddl.Store(time.Now().Add(i.expiration))
	return nil
}

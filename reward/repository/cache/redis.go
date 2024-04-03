package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/reward/domain"
)

type RewardRedisCache struct {
	client redis.Cmdable
}

func (c *RewardRedisCache) CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error {
	key := c.codeURLKey(r)
	data, err := json.Marshal(cu)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, time.Minute*29).Err()
}

func (c *RewardRedisCache) GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	key := c.codeURLKey(r)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.CodeURL{}, err
	}
	var res domain.CodeURL
	err = json.Unmarshal(data, &res)
	return res, err
}

func NewRewardRedisCache(client redis.Cmdable) RewardCache {
	return &RewardRedisCache{client: client}
}

func (c *RewardRedisCache) codeURLKey(r domain.Reward) string {
	return fmt.Sprintf("reward:code_url:%s:%d:%d", r.Target.Biz, r.Target.BizId, r.SrcUid)
}

package cache

import (
	"context"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"time"
)

type LocalCodeCache struct {
	cache      *lru.Cache
	lock       sync.Mutex
	expiration time.Duration
}

func NewLocalCodeCache(c *lru.Cache, expiration time.Duration) CodeCache {
	return &LocalCodeCache{
		cache:      c,
		expiration: expiration, // 过期时间初始化时传入
	}
}

func (lc *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	key := lc.key(biz, phone)
	now := time.Now()
	val, ok := lc.cache.Get(key)

	// 缓存中没有，表示从没发过
	if !ok {
		lc.cache.Add(key, codeItem{
			code:   code,
			cnt:    3,
			expire: now.Add(lc.expiration),
		})
		return nil
	}

	codeitm, ok := val.(codeItem)

	if !ok {
		return errors.New("系统错误")
	}

	// 还没过期，发太频繁了
	if now.Before(codeitm.expire) {
		return ErrCodeSendTooMany
	}

	// 有 firstKey 但过期了，重发
	lc.cache.Add(key, codeItem{
		code:   code,
		cnt:    3,
		expire: now.Add(lc.expiration),
	})
	return nil

}

func (lc *LocalCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	key := lc.key(biz, phone)
	val, ok := lc.cache.Get(key)
	if !ok {
		return false, ErrCodeKeyNotExist
	}

	codeitm, ok := val.(codeItem)

	if !ok {
		return false, errors.New("系统错误")
	}
	if codeitm.cnt <= 0 {
		// 为了安全，只提示验证太多次即可
		return false, ErrCodeVerifyTooMany
	}
	if code == codeitm.code {
		codeitm.cnt = -1
		return true, nil
	}
	codeitm.cnt -= 1
	return false, nil
}

func (lc *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type codeItem struct {
	code   string
	cnt    int
	expire time.Time
}

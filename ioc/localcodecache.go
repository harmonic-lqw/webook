package ioc

import (
	lru "github.com/hashicorp/golang-lru"
	"time"
)

func InitLRU() *lru.Cache {
	cache, err := lru.New(10)
	if err != nil {
		panic(nil)
	}
	return cache
}

func InitExpireTime() time.Duration {
	return time.Minute
}

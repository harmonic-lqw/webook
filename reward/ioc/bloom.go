package ioc

import (
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/bloom"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func InitBloomFilter() *bloom.Filter {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		panic(err)
	}
	RedisClient := redis.New(cfg.Addr)
	TestFilter := bloom.New(RedisClient, "BloomKey", 20*300)
	return TestFilter
}

package ioc

import (
	"github.com/redis/go-redis/v9"
	"webook/config"
)

func InitRedis() redis.Cmdable {
	// 压测要关闭
	// 限流，用 redis 来统计多实例上的总访问量
	return redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
}

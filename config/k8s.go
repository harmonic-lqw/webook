//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:123456@tcp(webook-mysql:3306)/webook",
	},
	Redis: RedisConfig{
		Addr: "webook-redis:6379",
	},
}

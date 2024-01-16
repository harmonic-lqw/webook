//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		InitRedis, ioc.InitDB,

		// Dao 和 Cache
		dao.NewUserDAO,
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		// LocalCodeCache
		//ioc.InitLRU,
		//ioc.InitExpireTime,
		//cache.NewRedisUserCache, cache.NewLocalCodeCache,

		// Repository
		repository.NewCachedUserRepository, repository.NewCachedCodeRepository,

		// Service
		InitSMSService, service.NewUserService, service.NewCodeService,

		// Handler
		web.NewUserHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

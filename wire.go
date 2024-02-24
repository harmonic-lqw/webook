//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"
)

// 纵向配置
var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,

		// Dao 和 Cache
		dao.NewUserDAO, dao.NewArticleGROMDAO,
		cache.NewRedisUserCache, cache.NewRedisCodeCache, cache.NewArticleRedisCache,
		ioc.InitConsumers,
		// LocalCodeCache
		//ioc.InitLRU,
		//ioc.InitExpireTime,
		//cache.NewRedisUserCache, cache.NewLocalCodeCache,

		// Repository
		repository.NewCachedUserRepository, repository.NewCachedCodeRepository, repository.NewCachedArticleRepository,

		// Service
		ioc.InitSMSService, ioc.InitWechatService, service.NewUserService, service.NewCodeService, service.NewArticleService,

		interactiveSvcSet,

		article.NewSaramaSyncProducer,
		article.NewInteractiveReadEventConsumer,
		//article.NewBatchInteractiveReadEventConsumer,

		// Handler
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}

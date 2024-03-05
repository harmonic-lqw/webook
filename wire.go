//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/interactive/events"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
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
var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

var rankingSvcSet = wire.NewSet(cache.NewRankingRedisCache,
	repository.NewCachedOnlyRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		ioc.InitRlockClient,

		// Dao 和 Cache
		dao.NewUserDAO, dao.NewArticleGORMDAO,
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

		// Intr Client
		ioc.InitIntrClient,
		// assignment 11
		//ioc.InitIntrRepositoryClient,

		// ranking
		rankingSvcSet,
		ioc.InitRankingJob,
		ioc.InitJobs,

		article.NewSaramaSyncProducer,
		events.NewInteractiveReadEventConsumer,
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

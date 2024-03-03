//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
	"webook/internal/events/article"
	"webook/internal/job"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/service/sms"
	"webook/internal/service/sms/async"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"
)

var thirdProvider = wire.NewSet(InitRedis, InitDB,
	InitLogger,
	InitSyncProducer,
	InitSaramaClient,
)
var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewRedisUserCache,
	repository.NewCachedUserRepository,
	service.NewUserService)

var articleSvcProvider = wire.NewSet(
	dao.NewArticleGORMDAO,
	article.NewSaramaSyncProducer,
	cache.NewArticleRedisCache,
	repository.NewCachedArticleRepository,
	service.NewArticleService)

var rankServiceProvider = wire.NewSet(
	service.NewBatchRankingService,
	repository.NewCachedOnlyRankingRepository,
	cache.NewRankingRedisCache,
)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
)

var jobProviderSet = wire.NewSet(
	service.NewCronJobService,
	repository.NewPreemptJobRepository,
	dao.NewGORMJobDAO)

//go:generate wire
func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		articleSvcProvider,
		interactiveSvcProvider,
		cache.NewRedisCodeCache,
		repository.NewCachedCodeRepository,
		// service 部分
		// 集成测试我们显式指定使用内存实现
		ioc.InitMemorySMSService,

		// 指定啥也不干的 wechat service
		InitWechatService,
		service.NewCodeService,
		// handler 部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		//web.NewObservabilityHandler,
		ijwt.NewRedisJWTHandler,

		// gin 的中间件
		ioc.InitGinMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	// 随便返回一个
	return gin.Default()
}

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(thirdProvider,
		userSvcProvider,
		interactiveSvcProvider,
		article.NewSaramaSyncProducer,
		cache.NewArticleRedisCache,
		//wire.InterfaceValue(new(article.ArticleDAO), dao),
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler)
	return new(web.ArticleHandler)
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil)
}

func InitAsyncSmsService(svc sms.Service) *async.Service {
	wire.Build(thirdProvider, repository.NewAsyncSmsRepository,
		dao.NewGORMAsyncSmsDAO,
		async.NewService,
	)
	return &async.Service{}
}

func InitRankingService() service.RankingService {
	wire.Build(thirdProvider,
		interactiveSvcProvider,
		articleSvcProvider,
		// 用不上这个 user repo，所以随便搞一个
		wire.InterfaceValue(new(repository.UserRepository),
			&repository.CachedUserRepository{}),
		rankServiceProvider)
	return &service.BatchRankingService{}
}

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service2.NewInteractiveService(nil)
}

func InitJobScheduler() *job.Scheduler {
	wire.Build(jobProviderSet, thirdProvider, job.NewScheduler)
	return &job.Scheduler{}
}

func InitJwtHdl() ijwt.Handler {
	wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}

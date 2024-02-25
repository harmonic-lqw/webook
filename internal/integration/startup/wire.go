//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
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

var thirdPartySet = wire.NewSet(
	// 第三方依赖
	InitRedis, InitDB, InitLogger,
	ioc.InitSaramaClient,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		// Dao 和 Cache
		dao.NewUserDAO, dao.NewArticleGROMDAO,
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		// LocalCodeCache
		//ioc.InitLRU,
		//ioc.InitExpireTime,
		//cache.NewRedisUserCache, cache.NewLocalCodeCache,

		// Repository
		repository.NewCachedUserRepository, repository.NewCachedCodeRepository, repository.NewCachedArticleRepository,

		// Service
		InitSMSService, service.NewUserService, service.NewCodeService, InitWechatService, service.NewArticleService,

		// Handler
		web.NewUserHandler, web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		//dao.NewArticleGROMDAO,
		service.NewArticleService,
		repository.NewCachedArticleRepository,
		article.NewSaramaSyncProducer,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}

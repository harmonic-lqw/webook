// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/google/wire"
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
	"webook/internal/web/jwt"
	"webook/ioc"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := ioc.InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, loggerV1)
	db := ioc.InitDB(loggerV1)
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, handler, codeService)
	wechatService := ioc.InitWechatService(loggerV1)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, handler, userService)
	articleDAO := dao.NewArticleGORMDAO(db)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, articleCache, userRepository)
	client := ioc.InitSaramaClient()
	syncProducer := ioc.InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	clientv3Client := ioc.InitEtcd()
	interactiveServiceClient := ioc.InitIntrClientV1(clientv3Client)
	articleHandler := web.NewArticleHandler(loggerV1, articleService, interactiveServiceClient)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler, loggerV1)
	v2 := ioc.InitConsumers()
	rankingCache := cache.NewRankingRedisCache(cmdable)
	rankingRepository := repository.NewCachedOnlyRankingRepository(rankingCache)
	rankingService := service.NewBatchRankingService(interactiveServiceClient, articleService, rankingRepository)
	rlockClient := ioc.InitRlockClient(cmdable)
	rankingJob := ioc.InitRankingJob(rankingService, loggerV1, rlockClient)
	cron := ioc.InitJobs(loggerV1, rankingJob)
	app := &App{
		server:    engine,
		consumers: v2,
		cron:      cron,
	}
	return app
}

// wire.go:

// 纵向配置
var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO, cache2.NewInteractiveRedisCache, repository2.NewCachedInteractiveRepository, service2.NewInteractiveService)

var rankingSvcSet = wire.NewSet(cache.NewRankingRedisCache, repository.NewCachedOnlyRankingRepository, service.NewBatchRankingService)

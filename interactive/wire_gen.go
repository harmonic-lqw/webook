// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/google/wire"
	"webook/interactive/events"
	"webook/interactive/grpc"
	"webook/interactive/ioc"
	"webook/interactive/repository"
	"webook/interactive/repository/cache"
	"webook/interactive/repository/dao"
	"webook/interactive/service"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitInteractiveAPP() *App {
	loggerV1 := ioc.InitLogger()
	srcDB := ioc.InitSrcDB(loggerV1)
	dstDB := ioc.InitDstDB(loggerV1)
	doubleWritePool := ioc.InitDoubleWritePool(srcDB, dstDB, loggerV1)
	db := ioc.InitBizDB(doubleWritePool)
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	cmdable := ioc.InitRedis()
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, interactiveCache, loggerV1)
	client := ioc.InitSaramaClient()
	interactiveReadEventConsumer := events.NewInteractiveReadEventConsumer(interactiveRepository, client, loggerV1)
	fixConsumer := ioc.InitFixerConsumer(client, loggerV1, srcDB, dstDB)
	v := ioc.InitConsumers(interactiveReadEventConsumer, fixConsumer)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	interactiveServiceServer := grpc.NewInteractiveServiceServer(interactiveService)
	server := ioc.NewGrpcxServer(interactiveServiceServer, loggerV1)
	syncProducer := ioc.InitSyncProducer(client)
	producer := ioc.InitInteractiveProducer(syncProducer)
	ginxServer := ioc.InitGinxServer(loggerV1, srcDB, dstDB, doubleWritePool, producer)
	app := &App{
		consumers:  v,
		server:     server,
		ginxServer: ginxServer,
	}
	return app
}

// wire.go:

var thirdPartySet = wire.NewSet(ioc.InitSrcDB, ioc.InitDstDB, ioc.InitBizDB, ioc.InitDoubleWritePool, ioc.InitRedis, ioc.InitLogger, ioc.InitSaramaClient, ioc.InitSyncProducer)

var interactiveSvcProvider = wire.NewSet(service.NewInteractiveService, repository.NewCachedInteractiveRepository, dao.NewGORMInteractiveDAO, cache.NewInteractiveRedisCache)

var migratorSarama = wire.NewSet(ioc.InitInteractiveProducer, ioc.InitFixerConsumer)

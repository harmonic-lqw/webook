//go:build wireinject

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

var thirdPartySet = wire.NewSet(
	//ioc.InitDB,
	ioc.InitSrcDB,
	ioc.InitDstDB,
	ioc.InitBizDB,
	ioc.InitDoubleWritePool,
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitSaramaClient,
	ioc.InitSyncProducer,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
)

var migratorSarama = wire.NewSet(
	ioc.InitInteractiveProducer,
	ioc.InitFixerConsumer,
)

func InitInteractiveAPP() *App {
	wire.Build(thirdPartySet, interactiveSvcProvider, migratorSarama,
		grpc.NewInteractiveServiceServer,
		//grpc.NewInteractiveRepositoryServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitGinxServer,
		ioc.InitConsumers,
		ioc.NewGrpcxServer,
		//ioc.NewGrpcxRepoServer,
		wire.Struct(new(App), "*"))
	return new(App)
}

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
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitSaramaClient,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
)

func InitInteractiveAPP() *App {
	wire.Build(thirdPartySet, interactiveSvcProvider,
		//grpc.NewInteractiveServiceServer,
		grpc.NewInteractiveRepositoryServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,
		//ioc.NewGrpcxServer,
		ioc.NewGrpcxRepoServer,
		wire.Struct(new(App), "*"))
	return new(App)
}

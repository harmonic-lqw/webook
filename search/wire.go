//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/search/events"
	"webook/search/grpc"
	"webook/search/ioc"
	"webook/search/repository"
	"webook/search/repository/dao"
	"webook/search/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewUserElasticDAO,
	dao.NewArticleElasticDAO,
	dao.NewAnyESDAO,
	dao.NewTagESDAO,
	repository.NewUserRepository,
	repository.NewArticleRepository,
	repository.NewAnyRepository,
	service.NewSyncService,
	service.NewSearchService,
)

var thirdProvider = wire.NewSet(
	ioc.InitESClient,
	ioc.InitEtcdClient,
	ioc.InitLogger,
	ioc.InitKafka)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		grpc.NewSyncServiceServer,
		grpc.NewSearchService,
		events.NewUserConsumer,
		events.NewArticleConsumer,
		ioc.InitGRPCxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

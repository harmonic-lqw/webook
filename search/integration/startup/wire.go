//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"webook/search/grpc"
	"webook/search/ioc"
	"webook/search/repository"
	"webook/search/repository/dao"
	"webook/search/service"
)

var serviceProviderSet = wire.NewSet(
	dao.NewUserElasticDAO,
	dao.NewArticleElasticDAO,
	dao.NewTagESDAO,
	dao.NewAnyESDAO,
	repository.NewUserRepository,
	repository.NewAnyRepository,
	repository.NewArticleRepository,
	service.NewSyncService,
	service.NewSearchService,
)

var thirdProvider = wire.NewSet(
	InitESClient,
	ioc.InitLogger)

func InitSearchServer() *grpc.SearchServiceServer {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		grpc.NewSearchService,
	)
	return new(grpc.SearchServiceServer)
}

func InitSyncServer() *grpc.SyncServiceServer {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		grpc.NewSyncServiceServer,
	)
	return new(grpc.SyncServiceServer)
}

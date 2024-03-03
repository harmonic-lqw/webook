//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"webook/interactive/grpc"
	"webook/interactive/repository"
	"webook/interactive/repository/cache"
	"webook/interactive/repository/dao"
	"webook/interactive/service"
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
)

var thirdProvider = wire.NewSet(InitRedis, InitDB,
	InitLogger,
)

//InitSyncProducer,
//InitSaramaClient,

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(thirdProvider, interactiveSvcProvider, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}

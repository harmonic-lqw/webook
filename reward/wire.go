//go:build wireinject

package reward

import (
	"github.com/google/wire"
	"webook/pkg/wego"
	"webook/reward/grpc"
	"webook/reward/ioc"
	"webook/reward/repository"
	"webook/reward/repository/cache"
	"webook/reward/repository/dao"
	"webook/reward/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitEtcdClient,
	ioc.InitRedis)

func Init() *wego.App {
	wire.Build(thirdPartySet,
		service.NewWechatNativeRewardService,
		ioc.InitAccountClient,
		ioc.InitGRPCServer,
		ioc.InitPaymentClient,
		repository.NewRewardRepository,
		cache.NewRewardRedisCache,
		dao.NewRewardGORMDAO,
		grpc.NewRewardServiceServer,

		ioc.InitBloomFilter,

		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}

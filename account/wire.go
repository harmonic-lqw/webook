package account

import (
	"github.com/google/wire"
	"webook/account/grpc"
	"webook/account/ioc"
	"webook/account/repository"
	"webook/account/repository/dao"
	"webook/account/service"
	"webook/pkg/wego"
)

func Init() *wego.App {
	wire.Build(
		ioc.InitDB,
		ioc.InitLogger,
		ioc.InitEtcdClient,
		ioc.InitGRPCxServer,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
		wire.Struct(new(wego.App), "GRPCServer"))
	return new(wego.App)
}

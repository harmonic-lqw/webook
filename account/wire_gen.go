// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package account

import (
	"webook/account/grpc"
	"webook/account/ioc"
	"webook/account/repository"
	"webook/account/repository/dao"
	"webook/account/service"
	"webook/pkg/wego"
)

// Injectors from wire.go:

func Init() *wego.App {
	db := ioc.InitDB()
	accountDAO := dao.NewCreditGORMDAO(db)
	accountRepository := repository.NewAccountRepository(accountDAO)
	accountService := service.NewAccountService(accountRepository)
	accountServiceServer := grpc.NewAccountServiceServer(accountService)
	client := ioc.InitEtcdClient()
	loggerV1 := ioc.InitLogger()
	server := ioc.InitGRPCxServer(accountServiceServer, client, loggerV1)
	app := &wego.App{
		GRPCServer: server,
	}
	return app
}

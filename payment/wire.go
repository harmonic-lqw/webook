//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/payment/grpc"
	"webook/payment/ioc"
	"webook/payment/repository"
	"webook/payment/repository/dao"
	"webook/payment/web"
	"webook/pkg/wego"
)

func InitApp() *wego.App {
	wire.Build(
		// 中间件
		ioc.InitEtcdClient,
		ioc.InitKafka,
		ioc.InitProducer,
		ioc.InitDB,
		ioc.InitLogger,

		// 微信支付服务
		ioc.InitWechatClient,
		ioc.InitWechatNativeService,
		ioc.InitWechatConfig,
		repository.NewPaymentRepository,
		dao.NewPaymentGORMDAO,

		// web
		ioc.InitWechatNotifyHandler,
		ioc.InitGinServer,
		web.NewWechatHandler,

		// 微服务
		grpc.NewWechatServiceServer,
		ioc.InitGRPCServer,

		// 定时任务
		ioc.InitScanLocalMessageJob,
		ioc.InitSyncWechatOrderJob,
		ioc.InitJobs,

		wire.Struct(new(wego.App), "WebServer", "GRPCServer", "Cron"),
	)
	return new(wego.App)
}

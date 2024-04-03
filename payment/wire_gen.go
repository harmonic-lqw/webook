// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"webook/payment/grpc"
	"webook/payment/ioc"
	"webook/payment/repository"
	"webook/payment/repository/dao"
	"webook/payment/web"
	"webook/pkg/wego"
)

// Injectors from wire.go:

func InitApp() *wego.App {
	wechatConfig := ioc.InitWechatConfig()
	handler := ioc.InitWechatNotifyHandler(wechatConfig)
	client := ioc.InitWechatClient(wechatConfig)
	db := ioc.InitDB()
	paymentDAO := dao.NewPaymentGORMDAO(db)
	paymentRepository := repository.NewPaymentRepository(paymentDAO)
	loggerV1 := ioc.InitLogger()
	saramaClient := ioc.InitKafka()
	producer := ioc.InitProducer(saramaClient)
	nativePaymentService := ioc.InitWechatNativeService(client, paymentRepository, loggerV1, producer, wechatConfig)
	wechatHandler := web.NewWechatHandler(handler, nativePaymentService, loggerV1)
	server := ioc.InitGinServer(wechatHandler)
	wechatServiceServer := grpc.NewWechatServiceServer(nativePaymentService)
	clientv3Client := ioc.InitEtcdClient()
	grpcxServer := ioc.InitGRPCServer(wechatServiceServer, clientv3Client, loggerV1)
	syncWechatOrderJob := ioc.InitSyncWechatOrderJob(nativePaymentService, loggerV1)
	scanLocalMessageJob := ioc.InitScanLocalMessageJob(nativePaymentService, producer, loggerV1)
	cron := ioc.InitJobs(loggerV1, syncWechatOrderJob, scanLocalMessageJob)
	app := &wego.App{
		WebServer:  server,
		GRPCServer: grpcxServer,
		Cron:       cron,
	}
	return app
}

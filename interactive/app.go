package main

import (
	"webook/interactive/events"
	"webook/pkg/ginx"
	"webook/pkg/grpcx"
)

type App struct {
	consumers  []events.Consumer
	server     *grpcx.Server
	ginxServer *ginx.Server
	// 一种启动服务器的写法
	//server *grpc.InteractiveServiceServer
}

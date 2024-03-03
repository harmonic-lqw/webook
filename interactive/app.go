package main

import (
	"webook/interactive/events"
	"webook/pkg/grpcx"
)

type App struct {
	consumers []events.Consumer
	server    *grpcx.Server

	// 一种启动服务器的写法
	//server *grpc.InteractiveServiceServer
}

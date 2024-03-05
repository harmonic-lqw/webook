package main

import (
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {
	initViper()

	app := InitInteractiveAPP()
	println(app)
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	// 开 数据迁移 服务
	go func() {
		er := app.ginxServer.Start()
		if er != nil {
			panic(er)
		}
	}()

	// 开 grpc 服务
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}

	// 一种启动服务器的写法
	//server := grpc.NewServer()
	//intrv1.RegisterInteractiveServiceServer(server, app.server)
	//l, err := net.Listen("tcp", ":8090")
	//if err != nil {
	//	panic(err)
	//}
	//_ = server.Serve(l)
}

func initViper() {
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

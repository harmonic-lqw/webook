package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViperWatch()
	app := InitApp()

	// 启动 grpc 服务
	go func() {
		err := app.GRPCServer.Serve()
		panic(err)
	}()

	app.Cron.Start()
	defer func() {
		ctx := app.Cron.Stop()
		<-ctx.Done()
	}()

	// 启动 web 服务
	err := app.WebServer.Start()
	panic(err)
}

func initViperWatch() {
	cfile := pflag.String("config", "config/dev.yaml", "配置文件路径")
	pflag.Parse()

	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

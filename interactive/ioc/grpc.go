package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	grpc2 "webook/interactive/grpc"
	"webook/pkg/grpcx"
	"webook/pkg/logger"
)

func NewGrpcxServer(intrSvc *grpc2.InteractiveServiceServer, l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port     int    `yaml:"port"`
		EtcdAddr string `yaml:"etcdAddr"`
		Name     string `yaml:"name"`
	}

	server := grpc.NewServer()
	// 反向注册
	intrSvc.Register(server)
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	return &grpcx.Server{
		Server:   server,
		EtcdAddr: cfg.EtcdAddr,
		Port:     cfg.Port,
		Name:     cfg.Name,
		L:        l,
	}
}

func NewGrpcxRepoServer(intrRepo *grpc2.InteractiveRepositoryServer) *grpcx.Server {
	server := grpc.NewServer()
	// 反向注册
	intrRepo.Register(server)
	return &grpcx.Server{
		Server: server,
	}
}

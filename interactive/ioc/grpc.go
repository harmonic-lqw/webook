package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	grpc2 "webook/interactive/grpc"
	"webook/pkg/grpcx"
)

func NewGrpcxServer(intrSvc *grpc2.InteractiveServiceServer) *grpcx.Server {
	server := grpc.NewServer()
	// 反向注册
	intrSvc.Register(server)
	addr := viper.GetString("grpc.server.addr")
	return &grpcx.Server{
		Server: server,
		Addr:   addr,
	}
}

func NewGrpcxRepoServer(intrRepo *grpc2.InteractiveRepositoryServer) *grpcx.Server {
	server := grpc.NewServer()
	// 反向注册
	intrRepo.Register(server)
	addr := viper.GetString("grpc.server.addr")
	return &grpcx.Server{
		Server: server,
		Addr:   addr,
	}
}

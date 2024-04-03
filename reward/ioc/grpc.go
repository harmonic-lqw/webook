package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"webook/pkg/grpcx"
	"webook/pkg/logger"
	grpcReward "webook/reward/grpc"
)

func InitGRPCServer(svc *grpcReward.RewardServiceServer, etcdClient *clientv3.Client, l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port    int   `yaml:"port"`
		EtcdTTL int64 `yaml:"etcdTTL"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	svc.Register(server)

	return &grpcx.Server{
		Server:  server,
		Port:    cfg.Port,
		Name:    "reward",
		L:       l,
		EtcdTTL: cfg.EtcdTTL,
		Client:  etcdClient,
	}

}

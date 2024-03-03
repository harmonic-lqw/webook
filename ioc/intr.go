package ioc

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	intrv1 "webook/api/proto/gen/intr/v1"
	intrv2 "webook/api/proto/gen/intr/v2"
	"webook/interactive/repository"
	"webook/interactive/service"
	"webook/internal/client"
)

func InitIntrClient(svc service.InteractiveService) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool
		Threshold int32 `yaml:"threshold"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}

	var opts []grpc.DialOption
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}
	remote := intrv1.NewInteractiveServiceClient(cc)
	local := client.NewLocalInteractiveServiceAdapter(svc)
	res := client.NewInteractiveClient(remote, local)
	viper.OnConfigChange(func(in fsnotify.Event) {
		cfg = Config{}
		err := viper.UnmarshalKey("grpc.client.intr", &cfg)
		if err != nil {
			panic(err)
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}

func InitIntrRepositoryClient(repo repository.InteractiveRepository) intrv2.InteractiveRepositoryClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool
		Threshold int32 `yaml:"threshold"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}

	var opts []grpc.DialOption
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}
	remote := intrv2.NewInteractiveRepositoryClient(cc)
	local := client.NewLocalInteractiveRepositoryClient(repo)
	res := client.NewInteractiveRepositoryClient(remote, local)
	viper.OnConfigChange(func(in fsnotify.Event) {
		cfg = Config{}
		err := viper.UnmarshalKey("grpc.client.intr", &cfg)
		if err != nil {
			panic(err)
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}

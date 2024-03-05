package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"webook/interactive/events"
	"webook/interactive/repository/dao"
	"webook/pkg/migrator/events/fixer"
)

func InitSaramaClient() sarama.Client {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(cfg.Addr, scfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}

func InitConsumers(c1 *events.InteractiveReadEventConsumer, c2 *fixer.FixConsumer[dao.Interactive]) []events.Consumer {
	return []events.Consumer{c1, c2}
}

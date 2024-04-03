package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"webook/payment/events"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	// Payment 的消息包含了支付-退款这种有顺序要求的事件，那么生产消息的时候就需要控制住有序性。
	// 在这里，我们在初始化生产者的时候指定使用哈希负载均衡算法
	saramaCfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner

	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitProducer(client sarama.Client) events.Producer {
	res, err := events.NewSaramaProducer(client)
	if err != nil {
		panic(err)
	}
	return res
}

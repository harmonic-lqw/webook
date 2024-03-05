package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"webook/interactive/repository/dao"
	"webook/pkg/ginx"
	"webook/pkg/gormx/connpool"
	"webook/pkg/logger"
	"webook/pkg/migrator/events"
	"webook/pkg/migrator/events/fixer"
	"webook/pkg/migrator/scheduler"
)

func InitGinxServer(l logger.LoggerV1,
	src SrcDB,
	dst DstDB,
	pool *connpool.DoubleWritePool,
	producer events.Producer,
) *ginx.Server {
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "harmonic",
		Subsystem: "webook_intr_admin",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	sch := scheduler.NewScheduler[dao.Interactive](l, src, dst, pool, producer)
	engine := gin.Default()
	sch.RegisterRoutes(engine)
	return &ginx.Server{
		Engine: engine,
		Addr:   viper.GetString("migrator.http.addr"),
	}
}

func InitInteractiveProducer(producer sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(producer, "inconsistent_interactive")
}

func InitFixerConsumer(client sarama.Client, l logger.LoggerV1, src SrcDB, dst DstDB) *fixer.FixConsumer[dao.Interactive] {
	consumer, err := fixer.NewFixConsumer[dao.Interactive](client, l, src, dst, "inconsistent_interactive")
	if err != nil {
		panic(err)
	}
	return consumer
}

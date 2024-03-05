package ioc

import (
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
	"webook/interactive/repository/dao"
	"webook/pkg/gormx"
	"webook/pkg/gormx/connpool"
	"webook/pkg/logger"
)

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitSrcDB(l logger.LoggerV1) SrcDB {
	return initDB(l, "src")
}

func InitDstDB(l logger.LoggerV1) DstDB {
	return initDB(l, "dst")
}

func InitBizDB(connPool *connpool.DoubleWritePool) *gorm.DB {
	doubleWriteDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: connPool,
	}))
	if err != nil {
		panic(err)
	}
	return doubleWriteDB
}

func initDB(l logger.LoggerV1, key string) *gorm.DB {
	type Config struct {
		DSN string
	}
	var cfg Config = Config{
		DSN: "root:123456@tcp(localhost:13316)/webook", // 默认值
	}
	err := viper.UnmarshalKey("db."+key, &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	//db, err := gorm.Open(mysql.Open(config.NoKSConfig.DB.DSN))
	if err != nil {
		panic(err)
	}

	// 接入prometheus
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook_" + key,
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	// 通过 callback 对数据库连接时间进行监控
	cb := gormx.NewCallbacks(prometheus2.SummaryOpts{
		Namespace: "harmonic",
		Subsystem: "webook",
		Name:      "gorm_db_" + key,
		Help:      "统计 GORM 的数据库查询",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	err = db.Use(cb)
	if err != nil {
		panic(err)
	}

	// GORM 中接入 OpenTelemetry
	err = db.Use(tracing.NewPlugin(
		tracing.WithoutMetrics(),
		tracing.WithDBName("webook"+key)))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
	g(s, logger.Field{Key: "args", Value: i})
}

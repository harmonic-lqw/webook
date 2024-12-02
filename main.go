package main

import (
	"bytes"
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log"
	"net/http"
	"time"
	"webook/internal/web/middleware"
	"webook/ioc"
)

func main() {
	initViperWatch()
	initPrometheus()

	tpCancel := ioc.InitOTEL()
	defer func() {
		// tp.Shutdown(ctx) 的调用需要清理资源等，因此本身也是耗时的，所以要控制住它
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		tpCancel(ctx)
	}()

	app := InitWebServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	app.cron.Start()
	defer func() {
		ctx := app.cron.Stop()
		<-ctx.Done()
	}()
	server := app.server

	//server.GET("/hello", func(ctx *gin.Context) {
	//	//	ctx.String(http.StatusOK, "成功部署")
	//	//})

	server.Run(":8080")
}

func initPrometheus() {
	go func() {
		// 专门给 Prometheus 开的端口，尽量不要和业务端口重合
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}

func useSession(server *gin.Engine) {
	// 基于 cookie 的实现
	// 存储数据的，也就是 userId 存的地方
	// 数据直接存储在 cookie 里
	store := cookie.NewStore([]byte("secret"))

	// 基于内存的实现
	// 第一个参数是 authentication key，第二个是 encryption key
	// Authentication：是指身份认证。
	// Encryption：是指数据加密。
	// 这两者再加上授权（权限控制），就是信息安全的三个核心概念。
	//store := memstore.NewStore([]byte("ezFwyB49AsChLgJIlqo2BxeUEOtS80tyorXuTt78bXslhPRZxTAlGIzUZWq0lU7X"),
	//	[]byte("kiRmG9M1YWzOojPNhIXtmrP52nPA1IfFCf5UG6TweEYhV79UdBKgzVbMRaLNAdug"))

	// 基于 Redis 的实现
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	//	[]byte("pBnSDaa0FCypBlPSpSoATWB4VZIS9niS"),
	//	[]byte("pBnSDaa0FCypBlPSpSoATWB4VZIS9niA"))
	//if err != nil {
	//	panic(err)
	//}

	// 初始化 session
	server.Use(sessions.Sessions("ssid", store))

	// 登录校验
	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(login.CheckLogin())
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

	val := viper.Get("test.key")
	log.Println(val)
}

func initViperV1() {
	// 设置默认值 default
	//viper.SetDefault("db.dsn", "root:123456@tcp(localhost:3316)/webook")

	viper.SetConfigFile("config/dev.yaml")
	viper.SetConfigType("yaml")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

// initViperV2 直接通过 viper.ReadConfig ，从字符串中读取
func initViperV2() {
	cfg := `
test:
  key: value1

redis:
  addr: "localhost:6379"

db:
  dsn: "root:123456@tcp(localhost:13316)/webook"`
	viper.SetConfigType("yaml")
	// 读取配置
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(err)
	}
}

// initViperV3 解决不同环境不同配置
func initViperV3() {
	cfgfile := pflag.String("config",
		"config/config.yaml", "配置文件路径")

	viper.SetConfigFile(*cfgfile)
	viper.SetConfigType("yaml")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

// initViperWatch 通过 WatchConfig 监听配置变更
func initViperWatch() {
	cfgfile := pflag.String("config",
		"config/dev.yaml", "配置文件路径")

	viper.SetConfigFile(*cfgfile)
	viper.SetConfigType("yaml")
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println(viper.Get("test.key"))
	})
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

// initViperRemote
func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")

	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}

	// viper 要注意并发安全问题
	go func() {
		for {
			err = viper.WatchRemoteConfig()
			if err != nil {
				panic(err)
			}
			log.Println("watch", viper.Get("test.key"))
			time.Sleep(time.Second * 5)
		}
	}()

}

package ioc

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"time"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/internal/web/middleware"
	"webook/pkg/ginx"
	prometheus2 "webook/pkg/ginx/middleware/prometheus"
	"webook/pkg/ginx/middleware/ratelimit"
	"webook/pkg/limiter"
	"webook/pkg/logger"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHdl *web.UserHandler, wechatHdl *web.OAuth2WechatHandler, artHdl *web.ArticleHandler, l logger.LoggerV1) *gin.Engine {
	ginx.SetLogger(l)
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	artHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable, hdl ijwt.Handler, l logger.LoggerV1) []gin.HandlerFunc {
	pb := &prometheus2.Builder{
		Namespace: "geektime_daming",
		Subsystem: "webook",
		Name:      "gin_http",
		Help:      "统计 GIN 的 HTTP 接口数据",
	}
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "http_biz_code",
		Help:      "统计业务错误码",
		ConstLabels: map[string]string{
			"instance_id": "my_instance_1",
		},
	})

	return []gin.HandlerFunc{
		cors.New(cors.Config{
			//AllowAllOrigins: true,
			//AllowOrigins:     []string{"http://localhost:3000"},
			AllowCredentials: true,
			AllowHeaders:     []string{"authorization", "content-type"},

			ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"}, // 允许前端访问后端响应中带的头部
			AllowOriginFunc: func(origin string) bool {
				return true
			},
			MaxAge: 12 * time.Hour,
		}),

		// 接入 Prometheus
		pb.BuildResponseTime(),
		pb.BuildActiveRequest(),

		// 接入 Opentelemetry
		otelgin.Middleware("webook"),

		// 限流
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),
		// 入口日志
		middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
			l.Debug("", logger.Field{Key: "req", Value: al})
		}).AllowReqBody().AllowRespBody().Builder(),
		// handler日志
		middleware.NewLogHandlerBuilder(l).Builder(),
		// 登录校验
		middleware.NewLoginJWTMiddlewareBuilder(hdl).CheckLogin(),
	}
}

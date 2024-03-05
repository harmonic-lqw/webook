package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"webook/pkg/logger"
)

var log = logger.NewNoOpLogger()

func SetLogger(l logger.LoggerV1) {
	log = l
}

var vector *prometheus.CounterVec

func InitCounter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
}

func WrapClaimsAndReq[Req any](fn func(*gin.Context, Req, UserClaims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("解析请求失败", logger.Error(err))
			return
		}
		// 因为泛型的限制，这里只能用包变量
		rawVal, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获得 claims",
				logger.String("path", ctx.Request.URL.Path))
			return
		}
		claims, ok := rawVal.(UserClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获得 claims",
				logger.String("path", ctx.Request.URL.Path))
			return
		}
		res, err := fn(ctx, req, claims)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			log.Error("执行业务逻辑失败",
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapReq[Req any](fn func(*gin.Context, Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("解析请求失败", logger.Error(err))
			return
		}
		res, err := fn(ctx, req)
		if err != nil {
			log.Error("执行业务逻辑失败",
				logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaims(fn func(*gin.Context, UserClaims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 因为泛型的限制，这里只能用包变量
		rawVal, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获得 claims",
				logger.String("path", ctx.Request.URL.Path))
			return
		}
		claims, ok := rawVal.(UserClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获得 claims",
				logger.String("path", ctx.Request.URL.Path))
			return
		}
		res, err := fn(ctx, claims)
		if err != nil {
			log.Error("执行业务逻辑失败",
				logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

func Wrap(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			log.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

package middleware

import (
	"github.com/gin-gonic/gin"
	"webook/pkg/logger"
)

type LogHandlerBuilder struct {
	logger logger.LoggerV1
}

func NewLogHandlerBuilder(logger logger.LoggerV1) *LogHandlerBuilder {
	return &LogHandlerBuilder{
		logger: logger,
	}
}

func (l *LogHandlerBuilder) Builder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		// 检查是否有错误发生
		if len(ctx.Errors) > 0 {
			// 有错误，记录日志
			l.logger.Error("", logger.Field{
				Key:   "logHandler",
				Value: ctx.Errors,
			})
		}
	}
}

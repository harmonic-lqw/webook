package middleware

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type LogMiddlewareBuilder struct {
	logFun        func(ctx context.Context, al AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddlewareBuilder(logFun func(ctx context.Context, al AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFun: logFun,
	}
}

func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddlewareBuilder) Builder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		if len(path) > 1024 {
			path = path[:1024]
		}
		method := ctx.Request.Method
		al := AccessLog{
			Path:   path,
			Method: method,
		}
		if l.allowReqBody && ctx.Request.Body != nil {
			// 忽略 err 不影响程序运行
			body, _ := ctx.GetRawData()
			// 因为 request.body 是一个 stream 流对象，因此只能读一次，所以这里读完要放回去
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			al.ReqBody = string(body)
		}

		// 如何拿到 response 里面的数据
		if l.allowRespBody {
			ctx.Writer = &responseWriter{
				ResponseWriter: ctx.Writer,
				al:             &al,
			}
		}

		defer func() {
			al.Duration = time.Since(start)
			l.logFun(ctx, al)
		}()

		ctx.Next()
	}
}

// AccessLog 具体一个日志，需要什么信息就添加什么
type AccessLog struct {
	Method   string        `json:"method"`
	Path     string        `json:"path"`
	ReqBody  string        `json:"req_body"`
	RespBody string        `json:"resp_body"`
	Status   int           `json:"status"`
	Duration time.Duration `json:"duration"`
}

// responseWriter 因为 gin 提供的 ResponseWriter 没有给出得到 data、statusCode 等的方法，因此为了能拿到 response 中的数据，使用装饰器自己定义一个扩展
type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.al.RespBody = string(data)
	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.al.Status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) WriteString(data string) (int, error) {
	rw.al.RespBody = data
	return rw.ResponseWriter.WriteString(data)
}

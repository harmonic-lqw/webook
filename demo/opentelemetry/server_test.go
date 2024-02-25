package opentelemetry

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"net/http"
	"testing"
	"time"
)

// 总结: opentelemetry 的使用
// 使用 opentelemetry 来打点要做的就是：
// 创建对应 span
// 记得关闭
// 调用 AddEvent 或 SetAttributes
// （在 zipkin 中：event 被转成了 Annotation；Attribute 被转成了 Tags）

func TestServer(t *testing.T) {
	res, err := newResource("demo", "v0.0.1")
	require.NoError(t, err)

	// 在客户端和服务端之间传递 tracing 的相关信息
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// 初始化 trace provider
	// 这个 provider 就是用来在打点的时候构建 trace 的
	tp, err := newTraceProvider(res)
	require.NoError(t, err)
	defer tp.Shutdown(context.Background())
	otel.SetTracerProvider(tp)

	// 创建 Span
	server := gin.Default()
	server.GET("/test", func(ginCtx *gin.Context) {
		// tracer 的名字，最好设置为唯一
		tracer := otel.Tracer("opentelemetry")
		var ctx context.Context = ginCtx

		ctx, span := tracer.Start(ctx, "top-span")
		defer span.End()
		time.Sleep(time.Second)
		// 强调整个流程中发生某事
		span.AddEvent("event1 happened")

		ctx, subSpan := tracer.Start(ctx, "sub-span")
		defer subSpan.End()
		time.Sleep(time.Millisecond * 300)
		// 强调整个流程上下文中有某些数据
		subSpan.SetAttributes(attribute.String("key1", "value1"))
		ginCtx.String(http.StatusOK, "OK")
	})
	server.Run("localhost:8082")
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion)))
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{})
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := zipkin.New(
		"http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			// Default is 5s, Set to 1s for demonstrative purposes
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}

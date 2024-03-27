package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"strings"
	"time"
	"webook/pkg/grpcx/interceptor"
)

type InterceptorBuilder struct {
	Namespace  string
	Subsystem  string
	Name       string
	InstanceId string
	Help       string
	interceptor.Builder
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	labels := []string{"type", "service", "method", "peer", "peer_ip", "code"}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      b.Name + "_resp_time",
		Help:      b.Help,
		ConstLabels: map[string]string{
			"instance_id": b.InstanceId,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(vector)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		defer func() {
			sv, method := b.splitMethodName(info.FullMethod)
			code := "OK"
			if err != nil {
				st, _ := status.FromError(err)
				code = st.Code().String()
			}
			duration := float64(time.Since(start).Milliseconds())
			vector.WithLabelValues("unary", sv, method, b.PeerName(ctx), b.PeerIP(ctx), code).Observe(duration)
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func (b *InterceptorBuilder) splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/")
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

package circuitbreaker

import (
	"context"
	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		err = b.breaker.Allow()
		if err == nil {
			resp, err = handler(ctx, req)
			if err == nil {
				b.breaker.MarkSuccess()
			} else {
				// 这里可以进行更加仔细的检测，只有代表服务端故障的 err，才 mark failed
				b.breaker.MarkFailed()
			}
			return
		} else {
			b.breaker.MarkFailed()
			return nil, status.Errorf(codes.Unavailable, "熔断")
		}

	}
}

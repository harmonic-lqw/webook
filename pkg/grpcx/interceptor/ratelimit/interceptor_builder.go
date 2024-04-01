package ratelimit

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"webook/pkg/limiter"
)

type InterceptorBuilder struct {
	limiter limiter.Limiter
	// 是根据 key 来做限流的
	key string
}

func NewInterceptorBuilder(limiter limiter.Limiter, key string) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key}
}

// BuildServerUnaryInterceptor 针对应用去限流, 此时 key 可以为 limiter:interactive-service
func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			// 保守做法：直接拒绝
			// 坏处的一个例子：如果 redis 崩了导致的 err，如果直接拒绝该请求，但其实该请求真正的处理代码没有用到 redis，就会有问题。因此这里可以做更详细的划分。
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
			// 激进做法：不用管，直接继续执行
		}

		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}
		return handler(ctx, req)
	}
}

// BuildServerUnaryInterceptorV1 标记降级
func (b *InterceptorBuilder) BuildServerUnaryInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		// 打一个标记位，然后在具体的业务中检测，判断是否降级
		if err != nil || limited {
			ctx = context.WithValue(ctx, "downgrade", "true")
		}

		return handler(ctx, req)
	}
}

// BuildServerUnaryInterceptorService 针对服务去限流, 此时 key 可以为 limiter:UserService
func (b *InterceptorBuilder) BuildServerUnaryInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if strings.HasPrefix(info.FullMethod, "/UserService") {
			limited, err := b.limiter.Limit(ctx, b.key)
			if err != nil {
				// 保守做法：直接拒绝
				// 坏处的一个例子：如果 redis 崩了导致的 err，如果直接拒绝该请求，但其实该请求真正的处理代码没有用到 redis，就会有问题。因此这里可以做更详细的划分。
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
				// 激进做法：不用管，直接继续执行
			}

			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
		}
		return handler(ctx, req)

	}
}

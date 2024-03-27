package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// BuildServerUnaryInterceptorBiz 针对具体业务去限流, 此时 key 可以为 limiter:user:get_by_id:<id>
func (b *InterceptorBuilder) BuildServerUnaryInterceptorBiz() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if getReq, ok := req.(*GetByIDRequest); ok {
			key := fmt.Sprintf("limiter:user:get_by_id:%d", getReq.Id)
			limited, err := b.limiter.Limit(ctx, key)
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

// LimiterUserServer 针对业务限流更推荐的做法是使用装饰器，只需要实现被限流的方法
type LimiterUserServer struct {
	limiter limiter.Limiter
	UserServiceServer
}

func (s *LimiterUserServer) GetByID(ctx context.Context, req *GetByIDRequest) (*GetByIDResponse, error) {
	key := fmt.Sprintf("limiter:user:get_by_id:%d", req.Id)
	limited, err := s.limiter.Limit(ctx, key)
	if err != nil {
		// 保守做法：直接拒绝
		// 坏处的一个例子：如果 redis 崩了导致的 err，如果直接拒绝该请求，但其实该请求真正的处理代码没有用到 redis，就会有问题。因此这里可以做更详细的划分。
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
		// 激进做法：不用管，直接继续执行
	}

	if limited {
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}

	return s.UserServiceServer.GetByID(ctx, req)

}

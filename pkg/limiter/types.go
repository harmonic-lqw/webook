package limiter

import "context"

type Limiter interface {
	// Limit 是否触发限流，返回 true 表示触发
	Limit(ctx context.Context, key string) (bool, error)
}

package grpc

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"sync/atomic"
	"time"
)

// CounterLimiter 收到请求的时候，计数器 +1；返回响应的时候，计数器-1
type CounterLimiter struct {
	cnt       atomic.Int32
	threshold int32
}

func (c *CounterLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 请求进来
		// 注意这里要先“占坑”，不要先判断是否超过阈值，目的是防止并发问题
		cnt := c.cnt.Add(1)
		defer func() {
			c.cnt.Add(-1)
		}()
		if cnt <= c.threshold {
			// 返回响应（这一步是真正发起了业务调用）
			resp, err = handler(ctx, req)
			return
		}

		return nil, status.Errorf(codes.ResourceExhausted, "计数器限流")
	}
}

// FixedWindowLimiter 将时间切成一个个窗口，确保每个窗口内的请求数量没有超过阈值
type FixedWindowLimiter struct {
	window          time.Duration
	lastWindowStart time.Time
	cnt             int
	threshold       int
	lock            sync.Mutex
}

func NewFixedWindowLimiter(window time.Duration, threshold int) *FixedWindowLimiter {
	return &FixedWindowLimiter{window: window, lastWindowStart: time.Now(), cnt: 0, threshold: threshold}
}

func (c *FixedWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		c.lock.Lock()

		// 检测是否超过固定窗口大小，如果超过，就开启一个新窗口
		now := time.Now()
		if now.After(c.lastWindowStart.Add(c.window)) {
			c.cnt = 0
			c.lastWindowStart = now
		}

		cnt := c.cnt + 1
		// 这里不用 defer 来解锁的原因：防止锁的范围过大，使 handler 串行运行
		c.lock.Unlock()
		if cnt < c.threshold {
			resp, err = handler(ctx, req)
			return
		}

		return nil, status.Errorf(codes.ResourceExhausted, "固定窗口限流")
	}
}

type SlidingWindowLimiter struct {
	window time.Duration
	// 关键问题：维持住每个请求的时间戳
	queue     queue.PriorityQueue[time.Time]
	threshold int
	lock      sync.Mutex
}

func (c *SlidingWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		c.lock.Lock()
		now := time.Now()

		// 快路径检测
		//if c.queue.Len() < c.threshold {
		//	_ = c.queue.Enqueue(now)
		//	c.lock.Unlock()
		//	resp, err = handler(ctx, req)
		//	return
		//}

		// 先去队列中查看是否时间戳都在滑动窗口范围内
		windowStart := now.Add(-c.window)
		for {
			first, _ := c.queue.Peek()
			if first.Before(windowStart) {
				// 不在，就把该时间戳删了
				_, _ = c.queue.Dequeue()
			} else {
				// 此时说明都在，因为这是一个优先级队列，时间戳最小的在前面
				break
			}
		}

		if c.queue.Len() < c.threshold {
			_ = c.queue.Enqueue(now)
			c.lock.Unlock()
			resp, err = handler(ctx, req)
			return
		}
		c.lock.Unlock()
		return nil, status.Errorf(codes.ResourceExhausted, "滑动窗口限流")
	}
}

type TokenBucketLimiter struct {
	// 产生令牌的频率
	interval time.Duration
	buckets  chan struct{}

	closeCh chan struct{}
	// 避免 closeCh 被多次关闭
	closeOnce sync.Once
}

func (c *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(c.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				select {
				case c.buckets <- struct{}{}:
				default:
					// 令牌桶满了，不用管，允许积压
				}
			case <-c.closeCh:
				return
			}
		}
	}()

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {

		select {
		case <-c.buckets:
			resp, err = handler(ctx, req)
			return
		// 做法1：直接返回错误
		default:
			return nil, status.Errorf(codes.ResourceExhausted, "令牌桶限流")
			// 做法2：阻塞，等待令牌，拿到令牌继续执行
			//case <-ctx.Done():
			//	return nil, ctx.Err()

		}
	}
}

func (c *TokenBucketLimiter) Close() error {
	c.closeOnce.Do(func() {
		close(c.closeCh)
	})
	return nil
}

type LeakyBucketLimiter struct {
	// 产生令牌的频率
	interval time.Duration
	closeCh  chan struct{}
	// 避免 closeCh 被多次关闭
	closeOnce sync.Once
}

func (c *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(c.interval)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {

		select {
		case <-ticker.C:
			resp, err = handler(ctx, req)
			return
		case <-c.closeCh:
			// 限流器关了。此时可以返回错误，也可以处理业务
			return nil, status.Errorf(codes.ResourceExhausted, "漏桶限流")

		// 做法1：直接返回错误
		default:
			return nil, status.Errorf(codes.ResourceExhausted, "漏桶限流")
			// 做法2：阻塞，等待令牌，拿到令牌继续执行
			//case <-ctx.Done():
			//	return nil, ctx.Err()
		}
	}
}

func (c *LeakyBucketLimiter) Close() error {
	c.closeOnce.Do(func() {
		close(c.closeCh)
	})
	return nil
}

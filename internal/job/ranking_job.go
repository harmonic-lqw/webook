package job

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
	"webook/internal/service"
	"webook/pkg/logger"
)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration
	client  *rlock.Client
	key     string
	l       logger.LoggerV1

	// 用本地锁保护分布式锁：抢锁和续约可能导致并发问题(r.lock == nil;r.lock = lock;r.lock = nil)
	localLock *sync.Mutex
	// 为了扩大锁的范围，将 lock 设置为一个字段，并通过本地锁保护起来
	lock *rlock.Lock
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, l logger.LoggerV1, client *rlock.Client) *RankingJob {
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		key:       "job:ranking",
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// Run 扩大锁的范围：实现只有一个实例负责热榜计算
func (r *RankingJob) Run() error {
	r.localLock.Lock()
	qLock := r.lock

	if qLock == nil {
		// 开始抢锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		// r.timeout 分布式锁持有时间;最长计算时间
		lock, err := r.client.Lock(ctx, r.key, r.timeout,
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
				// 重试的超时
			}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.lock = lock
		r.localLock.Unlock()
		// 续约机制！
		go func() {
			// 这是一个阻塞调用，会一致在这里续约
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			// 续约失败
			if er != nil {
				// 没法中断当下正在调度的热榜计算（如果有）
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

// Run()只用分布式锁来控制计算
//func (r *RankingJob) Run() error {
//	// 加锁超时，比所有重试时间要长一些
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
//	defer cancel()
//
//	// r.timeout 分布式锁持有时间;最长计算时间
//	lock, err := r.client.Lock(ctx, r.key, r.timeout,
//		&rlock.FixIntervalRetry{
//			Interval: time.Millisecond * 100,
//			Max:      3,
//			// 重试的超时
//		}, time.Second)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		// 解锁超时
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//		defer cancel()
//		er := lock.Unlock(ctx)
//		if er != nil {
//			// 主动释放分布式锁失败没有关系，也不需要重试，因为超时(r.timeout)后会自动释放
//			r.l.Error("ranking job释放分布式锁失败",
//				logger.Error(er))
//		}
//	}()
//	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
//	defer cancel()
//	return r.svc.TopN(ctx)
//}

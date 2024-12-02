package job

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/shirou/gopsutil/mem"
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

	// 模拟负载
	nodeId int64
	load   int
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, l logger.LoggerV1, client *rlock.Client, nodeId int64, load int) *RankingJob {
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		key:       "job:ranking",
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
		nodeId:    nodeId,
		load:      load,
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
		// Week11作业：无锁时，如果负载过高就放弃抢锁
		if r.load >= 90 {
			return nil
		}

		// Week11作业：判断该节点是否在所有节点中属于低负载节点
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_, minLoad, err := r.svc.GetMinLoadNode(ctx)
		if err != nil {
			r.l.Error("获取最小负载节点失败",
				logger.Error(err))
			return err
		}

		// Week11作业：当负载较高(>50) 且 当前节点比全局最小节点负载大太多(20)时，同样放弃抢锁
		if r.load > 50 && (r.load-minLoad) > 20 {
			return nil
		}

		// 开始抢锁
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
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

	// Week11作业：有锁时，负载过高释放锁（拿到锁也再次判断一次）
	if r.load >= 90 {
		r.localLock.Lock()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := r.lock.Unlock(ctx)
		if err != nil {
			r.l.Error("负载过高，但释放分布式锁失败",
				logger.Error(err))
			return err
		}
		r.localLock.Unlock()
		// 或许不用返回，多计算一次也没啥大不了
		return nil
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

// RefreshLoad 刷新负载
func (r *RankingJob) RefreshLoad() {
	//r.load = rand.Intn(100)

	// 统计内存使用率作为负载
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}
	r.load = int(memInfo.UsedPercent)
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			//r.load = rand.Intn(100)
			memInfo, err = mem.VirtualMemory()
			if err != nil {
				// 此时继续沿用上一次的统计量
				r.l.Warn("计算内存使用率失败", logger.Error(err))
			} else {
				r.load = int(memInfo.UsedPercent)
			}

			// 上传负载
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := r.svc.SetLoad(ctx, r.nodeId, r.load)
			if err != nil {
				r.l.Warn("上传该节点负载失败",
					logger.Int64("nodeId", r.nodeId),
					logger.Int("load", r.load))
			}
			cancel()
		}
	}()
}

// Run 只用分布式锁来控制计算
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

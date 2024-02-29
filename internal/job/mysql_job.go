package job

import (
	"context"
	"golang.org/x/sync/semaphore"
	"time"
	"webook/internal/service"
	"webook/pkg/logger"
)

type Scheduler struct {
	dbTimeout time.Duration
	svc       service.CronJobService
	executors map[string]Executor
	l         logger.LoggerV1

	// 通过信号量，控制可以调度的任务数量
	limiter *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{svc: svc,
		dbTimeout: time.Second * 5,
		limiter:   semaphore.NewWeighted(100),
		executors: map[string]Executor{},
		l:         l,
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return
		}

		// 抢占
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		// 运行1
		exec, ok := s.executors[j.Executor]
		if !ok {
			// 属于编程错误
			s.l.Error("找不到执行器",
				logger.Int64("jid", j.Id),
				logger.String("executor", j.Executor))
			continue
		}

		go func() {
			// 释放
			defer func() {
				s.limiter.Release(1)
				j.CancelFunc()
			}()

			// 运行2
			er := exec.Exec(ctx, j)
			if er != nil {
				s.l.Error("执行任务失败",
					logger.Int64("jid", j.Id),
					logger.Error(er))
				return
			}

			// 重置下一次调度时间
			er = s.svc.ResetNextTime(ctx, j)
			if er != nil {
				s.l.Error("重置下次执行时间失败",
					logger.Int64("jid", j.Id),
					logger.Error(er))
			}
		}()

	}
}

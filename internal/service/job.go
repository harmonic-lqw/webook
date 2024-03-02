package service

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error

	// 可以实现对 job 增删改查的方法
}

type cronJobService struct {
	repo repository.CronJobRepository
	l    logger.LoggerV1

	// 续约间隔
	refreshInterval time.Duration
}

func NewCronJobService(repo repository.CronJobRepository, l logger.LoggerV1) CronJobService {
	return &cronJobService{repo: repo,
		l:               l,
		refreshInterval: time.Minute}
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}

	// 续约
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()

	j.CancelFunc = func() {
		// 取消续约
		ticker.Stop()
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.repo.Release(newCtx, j.Id)
		if er != nil {
			c.l.Error("释放 job 失败",
				logger.Error(er),
				logger.Int64("jid", j.Id))
		}
	}
	return j, nil
}

func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.UpdateNextTime(ctx, j.Id, nextTime)
}

func (c *cronJobService) refresh(jid int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, jid)
	if err != nil {
		c.l.Error("续约失败",
			logger.Error(err),
			logger.Int64("jid", jid))
	}
}

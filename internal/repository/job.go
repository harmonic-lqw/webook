package repository

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error
}

type PreemptJobRepository struct {
	dao dao.CronJobDAO
}

func NewPreemptJobRepository(dao dao.CronJobDAO) CronJobRepository {
	return &PreemptJobRepository{
		dao: dao,
	}
}

func (p *PreemptJobRepository) UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error {
	return p.dao.UpdateNextTime(ctx, jid, nextTime)
}

func (p *PreemptJobRepository) UpdateUtime(ctx context.Context, jid int64) error {
	return p.dao.UpdateUtime(ctx, jid)
}

func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	return domain.Job{
		Id:         j.Id,
		Name:       j.Name,
		Executor:   j.Executor,
		Expression: j.Expression,
		Cfg:        j.Cfg,
	}, err

}

func (p *PreemptJobRepository) Release(ctx context.Context, jid int64) error {
	return p.dao.Release(ctx, jid)
}

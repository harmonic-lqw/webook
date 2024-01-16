package repository

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

type AsyncSmsRepository interface {
	// Add 添加一个异步 SMS 记录
	Add(ctx context.Context, s domain.AsyncSms) error
	// PreemptWaitingSMS 抢占式获取一个待发送短信
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error)
	// ReportScheduleResult 根据是否发送成功来标记数据库中，短信的状态
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

type asyncSmsRepository struct {
	dao dao.AsyncSmsDAO
}

func NewAsyncSmsRepository(dao dao.AsyncSmsDAO) AsyncSmsRepository {
	return &asyncSmsRepository{
		dao: dao,
	}
}

func (a *asyncSmsRepository) Add(ctx context.Context, s domain.AsyncSms) error {
	return a.dao.Insert(ctx, dao.AsyncSms{
		Config: sqlx.JsonColumn[dao.SmsConfig]{
			Val: dao.SmsConfig{
				TplId:   s.TplId,
				Args:    s.Args,
				Numbers: s.Numbers,
			},
			Valid: true,
		},
		RetryMax: s.RetryMax,
	})
}

func (a *asyncSmsRepository) PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error) {
	as, err := a.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.AsyncSms{}, err
	}
	return a.toDomain(as), nil
}

func (a *asyncSmsRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return a.dao.MarkSuccess(ctx, id)
	}
	return a.dao.MarkFailed(ctx, id)
}

func (a *asyncSmsRepository) toDomain(as dao.AsyncSms) domain.AsyncSms {
	return domain.AsyncSms{
		Id:       as.Id,
		TplId:    as.Config.Val.TplId,
		Args:     as.Config.Val.Args,
		Numbers:  as.Config.Val.Numbers,
		RetryMax: as.RetryMax,
	}
}

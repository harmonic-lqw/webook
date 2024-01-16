package dao

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

type AsyncSmsDAO interface {
	Insert(ctx context.Context, s AsyncSms) error
	// GetWaitingSMS 获取一个待发送的短信，注意并发问题和重试策略
	GetWaitingSMS(ctx context.Context) (AsyncSms, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

const (
	// 等待发送
	asyncStatusWaiting = iota
	// 发送失败（已经超过了重试次数）
	asyncStatusFailed
	// 已经发送成功
	asyncStatusSuccess
)

type GORMAsyncSmsDAO struct {
	db *gorm.DB
}

func (g *GORMAsyncSmsDAO) Insert(ctx context.Context, s AsyncSms) error {
	return g.db.Create(&s).Error
}

func (g *GORMAsyncSmsDAO) GetWaitingSMS(ctx context.Context) (AsyncSms, error) {
	// 对高并发场景不太适用，因为 select 对数据库的压力会很大
	var s AsyncSms
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 这里使用抢占式的策略，我们获取一分钟之前的异步短信发送请求，在获取到之后立即更新时间
		// 这个设计非常巧妙，一方面防止了多个部署的实例拿到这个请求，同时又实现了间隔重试
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("utime < ? and status = ?", endTime, asyncStatusWaiting).First(&s).Error
		if err != nil {
			return err
		}

		// 只要更新了更新时间，根据前面的筛选规则，就不可能被别的节点抢占了
		err = tx.Model(&AsyncSms{}).
			Where("id = ?", s.Id).
			Updates(map[string]any{
				"retry_cnt": gorm.Expr("retry_cnt + 1"),
				// 更新时间，如果发送失败，相当于一分钟后重试
				"utime": now,
			}).Error
		return err
	})
	return s, err
}

func (g *GORMAsyncSmsDAO) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&AsyncSms{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusSuccess,
		}).Error
}

func (g *GORMAsyncSmsDAO) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&AsyncSms{}).
		Where("id = ? and retry_cnt >= retry_max", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusFailed,
		}).Error
}

func NewGORMAsyncSmsDAO(db *gorm.DB) AsyncSmsDAO {
	return &GORMAsyncSmsDAO{
		db: db,
	}
}

type AsyncSms struct {
	Id     int64
	Config sqlx.JsonColumn[SmsConfig]
	// 重试次数
	RetryCnt int
	// 最大次数
	RetryMax int
	Status   uint8
	Ctime    int64
	Utime    int64 `gorm:"index"`
}

type SmsConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}

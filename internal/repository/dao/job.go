package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type CronJobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGORMJobDAO(db *gorm.DB) CronJobDAO {
	return &GORMJobDAO{db: db}
}

// Preempt 抢占核心代码
func (dao *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := dao.db.WithContext(ctx)
	for {
		var j Job
		now := time.Now()
		err := db.Where("(status = ? AND next_time < ?) OR (status = ? AND utime < ?)",
			jobStatusWaiting, now.UnixMilli(),
			// utime 的判定需要根据 cronJobService.refreshInterval 续约间隔来决定
			jobStatusRunning, now.Add(-time.Minute).UnixMilli()).
			First(&j).Error
		if err != nil {
			return j, err
		}
		now = time.Now()
		res := db.Model(&Job{}).
			Where("id = ? AND version = ?", j.Id, j.Version).
			Updates(map[string]any{
				"status":  jobStatusRunning,
				"version": j.Version + 1,
				"utime":   now.UnixMilli(),
			})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 没抢到继续抢
			continue
		}
		return j, err
	}
}

func (dao *GORMJobDAO) Release(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).
		Updates(map[string]any{
			"status": jobStatusWaiting,
			"utime":  now,
		}).Error

}

func (dao *GORMJobDAO) UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).Updates(map[string]any{
		"utime":     now,
		"next_time": nextTime.UnixMilli(),
	}).Error
}

func (dao *GORMJobDAO) UpdateUtime(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).
		Updates(map[string]any{
			"utime": now,
		}).Error
}

type Job struct {
	Id       int64  `gorm:"primaryKey, autoIncrement"`
	Name     string `gorm:"type=varchar(128), unique"`
	Executor string

	// 一些配置信息
	Cfg string

	// cron 表达式
	Expression string

	// 状态来表达，是不是可以被抢占，是否已经被抢占
	Status int

	// 乐观锁，体会 Version 的用法
	Version int

	// 用于判断是否到需要被执行的时间
	NextTime int64 `gorm:"index"`

	Ctime int64
	Utime int64
}

const (
	// jobStatusWaiting 没人抢占
	jobStatusWaiting = iota
	// jobStatusRunning 已经被人抢占
	jobStatusRunning
	// jobStatusPaused 停止调度
	jobStatusPaused
)

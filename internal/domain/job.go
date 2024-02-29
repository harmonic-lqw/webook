package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type Job struct {
	Id int64
	// executor 通过 Executor/Name 可以确定该 Job 的执行方法
	Name     string // executor 中 exec func 的名字
	Executor string // executor 的名字
	// cron 表达式
	Expression string

	Cfg        string
	CancelFunc func()
}

func (j Job) NextTime() time.Time {
	c := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	s, _ := c.Parse(j.Expression)
	return s.Next(time.Now())
}

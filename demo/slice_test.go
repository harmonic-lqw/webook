package demo

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestCronJobs(t *testing.T) {
	// 创建 cron 实例
	c := cron.New()

	// 添加第一个定时任务
	_, err := c.AddJob("@every 1s", Job1(func() {
		t.Log("Task 1 executed")
	}))
	assert.NoError(t, err)

	// 添加第二个定时任务
	_, err = c.AddJob("@every 5s", Job2(func() {
		t.Log("Task 2 executed")
	}))
	assert.NoError(t, err)

	c.Start() // 调用 Start 才开始调度任务
	time.Sleep(time.Second * 20)
	// 这两步才算停下来
	ctx := c.Stop() // 代表暂停调度，同时不能再调度新任务，但是正在执行的可以继续执行
	t.Log("发出停止信号")
	<-ctx.Done() // 代表所有任务执行完毕，真正退出了
	t.Log("彻底停下来了")

}

type Job1 func()
type Job2 func()

func (j1 Job1) Run() {
	j1()
}

func (j2 Job2) Run() {
	j2()
}

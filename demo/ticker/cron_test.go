package ticker

import (
	cron "github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCronExpr(t *testing.T) {
	expr := cron.New(cron.WithSeconds())

	id, err := expr.AddFunc("@every 1s", func() {
		t.Log("执行")
	})
	assert.NoError(t, err)
	t.Log("任务", id)
	expr.Start() // 调用 Start 才开始调度任务
	time.Sleep(time.Second * 5)
	// 这两步才算停下来
	ctx := expr.Stop() // 代表暂停调度，同时不能再调度新任务，但是正在执行的可以继续执行
	t.Log("发出停止信号")
	<-ctx.Done() // 代表所有任务执行完毕，真正退出了
	t.Log("彻底停下来了")

}

// go 中较为常见的使用函数衍生类型来实现单一接口
type JobFunc func()

func (j JobFunc) Run() {
	j()
}

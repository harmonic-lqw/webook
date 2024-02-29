package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"strconv"
	"time"
	"webook/pkg/logger"
)

// CronJobBuilder 为什么要有这个CronJobBuilder
// 因为：cron.New(cron.WithSeconds()).AddJob()，需要添加 cron.Job ，因此需要 Build 转化，其中还可以进行监控等操作
type CronJobBuilder struct {
	l      logger.LoggerV1
	vector *prometheus.SummaryVec
}

func NewCronJobBuilder(l logger.LoggerV1, opts prometheus.SummaryOpts) *CronJobBuilder {
	vector := prometheus.NewSummaryVec(opts, []string{"job", "success"})
	return &CronJobBuilder{l: l, vector: vector}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdapterFunc(func() {
		start := time.Now()
		b.l.Debug("开始运行",
			logger.String("name", name))

		err := job.Run()

		if err != nil {
			b.l.Error("执行失败",
				logger.Error(err),
				logger.String("name", name))
		}

		b.l.Debug("结束运行",
			logger.String("name", name))
		duration := time.Since(start).Milliseconds()
		b.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).Observe(float64(duration))
	})
}

type cronJobAdapterFunc func()

func (c cronJobAdapterFunc) Run() {
	c()
}

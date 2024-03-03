package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"time"
	"webook/internal/job"
	"webook/internal/service"
	"webook/pkg/logger"
)

func InitRankingJob(svc service.RankingService, l logger.LoggerV1, client *rlock.Client) *job.RankingJob {
	// 这里的时间是指，计算一次 topN 最多花多长时间
	rankingJob := job.NewRankingJob(svc, time.Second*30, l, client, 1, 0)
	rankingJob.RefreshLoad()
	return rankingJob
}

func InitJobs(l logger.LoggerV1, rJob *job.RankingJob) *cron.Cron {
	builder := job.NewCronJobBuilder(l, prometheus.SummaryOpts{
		Namespace: "harmonic",
		Subsystem: "webook",
		Name:      "cron_job",
		Help:      "定时任务执行",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	expr := cron.New(cron.WithSeconds())
	// 定时任务 1m 执行一次计算 topN
	_, err := expr.AddJob("@every 1m", builder.Build(rJob))
	if err != nil {
		panic(err)
	}
	return expr
}

package cronjobx

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"time"
	"webook/pkg/logger"
)

type CronJobBuilder struct {
	l      logger.LoggerV1
	vector *prometheus.SummaryVec
	tracer trace.Tracer
}

func NewCronJobBuilder(l logger.LoggerV1) *CronJobBuilder {
	p := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "harmonic",
		Subsystem: "webook",
		Help:      "统计定时任务执行情况",
		Name:      "cron_job",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"name", "success"})
	prometheus.MustRegister(p)
	return &CronJobBuilder{
		l:      l,
		vector: p,
		tracer: otel.GetTracerProvider().Tracer("webook/internal/job"),
	}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdapterFunc(func() error {
		_, span := b.tracer.Start(context.Background(), name)
		defer span.End()
		start := time.Now()
		b.l.Info("任务开始",
			logger.String("job", name))
		var success bool
		defer func() {
			b.l.Info("任务结束",
				logger.String("job", name))
			duration := time.Since(start).Milliseconds()
			b.vector.WithLabelValues(name, strconv.FormatBool(success)).Observe(float64(duration))
		}()
		err := job.Run()
		success = err == nil
		if err != nil {
			span.RecordError(err)
			b.l.Error("运行任务失败", logger.Error(err),
				logger.String("job", name))
		}
		return nil
	})
}

type cronJobAdapterFunc func() error

func (c cronJobAdapterFunc) Run() {
	_ = c()
}

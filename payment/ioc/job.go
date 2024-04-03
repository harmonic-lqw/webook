package ioc

import (
	"github.com/robfig/cron/v3"
	"webook/payment/events"
	"webook/payment/job"
	"webook/payment/service/wechat"
	"webook/pkg/cronjobx"
	"webook/pkg/logger"
)

func InitSyncWechatOrderJob(svc *wechat.NativePaymentService, l logger.LoggerV1) *job.SyncWechatOrderJob {
	return job.NewSyncWechatOrderJob(svc, l)
}

func InitScanLocalMessageJob(svc *wechat.NativePaymentService, producer events.Producer, l logger.LoggerV1) *job.ScanLocalMessageJob {
	return job.NewScanLocalMessageJob(svc, producer, l)
}

func InitJobs(l logger.LoggerV1, orderJob *job.SyncWechatOrderJob, messageJob *job.ScanLocalMessageJob) *cron.Cron {
	builder := cronjobx.NewCronJobBuilder(l)
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddJob("@every 10m", builder.Build(orderJob))
	if err != nil {
		panic(err)
	}
	_, err = expr.AddJob("@every 5m", builder.Build(messageJob))
	if err != nil {
		panic(err)
	}
	return expr
}

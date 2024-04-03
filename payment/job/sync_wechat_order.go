package job

import (
	"context"
	"time"
	"webook/payment/service/wechat"
	"webook/pkg/logger"
)

type SyncWechatOrderJob struct {
	svc *wechat.NativePaymentService
	l   logger.LoggerV1
}

func NewSyncWechatOrderJob(svc *wechat.NativePaymentService, l logger.LoggerV1) *SyncWechatOrderJob {
	return &SyncWechatOrderJob{
		svc: svc,
		l:   l,
	}
}

func (s *SyncWechatOrderJob) Name() string {
	return "sync_wechat_order_job"
}

func (s *SyncWechatOrderJob) Run() error {
	offset := 0
	const limit = 100
	// 查找 30 分钟以前的订单，是因为调用 prepay 时设置了二维码过期时间 30 分钟
	now := time.Now().Add(-time.Minute * 31)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		pmts, err := s.svc.FindExpiredPayment(ctx, offset, limit, now)
		cancel()
		if err != nil {
			// 直接中断此次任务执行
			return err
		}
		// 因为微信没有提供批量接口，所以我们这里只能单个查询
		for _, pmt := range pmts {
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			err = s.svc.SyncWechatInfo(ctx, pmt.BizTradeNO)
			if err != nil {
				// 也可以中断，也可以只记录日志
				s.l.Error("同步微信支付信息失败",
					logger.String("trade_no", pmt.BizTradeNO),
					logger.Error(err))
			}
			cancel()
		}
		if len(pmts) < limit {
			return nil
		}
		offset = offset + limit
	}
}

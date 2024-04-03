package job

import (
	"context"
	"time"
	"webook/payment/events"
	"webook/payment/service/wechat"
	"webook/pkg/logger"
)

// ScanLocalMessageJob 定时从本地消息表中取消息然后发送
type ScanLocalMessageJob struct {
	svc      *wechat.NativePaymentService
	producer events.Producer
	l        logger.LoggerV1
}

func NewScanLocalMessageJob(svc *wechat.NativePaymentService, producer events.Producer, l logger.LoggerV1) *ScanLocalMessageJob {
	return &ScanLocalMessageJob{svc: svc, producer: producer, l: l}
}

func (s *ScanLocalMessageJob) Name() string {
	return "scan_local_message"
}

func (s *ScanLocalMessageJob) Run() error {
	offset := 0
	limit := 100
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		// 查找未发送的消息
		msgs, err := s.svc.FindMessage(ctx, offset, limit)
		cancel()
		if err != nil {
			return err
		}
		for _, msg := range msgs {
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			err = s.producer.ProducePaymentEvent(ctx, events.PaymentEvent{
				BizTradeNO: msg.BizTradeNO,
				Status:     msg.Status,
			})
			cancel()
			if err != nil {
				continue
			}
			// 如果发送成功就更新数据库中消息状态
			err = s.svc.UpdateMessageById(ctx, msg.Id)
			if err != nil {
				continue
			}
		}
	}
}

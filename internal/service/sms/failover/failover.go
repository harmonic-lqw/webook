package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"webook/internal/service/sms"
)

type FailOverSMSService struct {
	svcs []sms.Service

	// sendV1 字段
	idx uint64
}

func NewFailOverSMSService(svcs []sms.Service) *FailOverSMSService {
	return &FailOverSMSService{
		svcs: svcs,
	}
}

func (f *FailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		log.Println(err)
	}
	return errors.New("轮询所有服务商，均发送失败")
}

// SendV1 改进 Send 中，轮询的起点，大部分都打在 svc0 处
// 但这里并不是严格轮询，而只是对起始的 svc 进行轮询
func (f *FailOverSMSService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[i%length]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.Canceled, context.DeadlineExceeded:
			// 前者是被取消 后者是超时，两者都是和用户体验密切相关的错误，因此这里进行返回
			return err
		}
		log.Println(err)
	}
	return errors.New("轮询所有服务商，均发送失败")

}

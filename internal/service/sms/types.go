package sms

import "context"

// Service 发送短信的抽象
// 用于屏蔽不同供应商发送短信的区别
type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}

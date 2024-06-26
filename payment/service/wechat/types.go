package wechat

import (
	"context"
	"webook/payment/domain"
)

type PaymentService interface {
	// Prepay 预支付，用于微信创建订单
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}

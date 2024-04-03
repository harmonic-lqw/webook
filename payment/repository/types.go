package repository

import (
	"context"
	"time"
	"webook/payment/domain"
)

type PaymentRepository interface {
	AddPayment(ctx context.Context, pmt domain.Payment) error
	GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error)
	UpdatePayment(ctx context.Context, pmt domain.Payment) error
	FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error)

	// assignment week17
	CreateLocalMessage(ctx context.Context, bizTradeNO string, status domain.PaymentStatus) error
	FindMessage(ctx context.Context, offset int, limit int) ([]domain.PaymentMessage, error)
	UpdateMessageById(ctx context.Context, id int64) error
}

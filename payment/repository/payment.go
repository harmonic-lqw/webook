package repository

import (
	"context"
	"time"
	"webook/payment/domain"
	"webook/payment/repository/dao"
)

type paymentRepository struct {
	dao dao.PaymentDAO
}

func (p *paymentRepository) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error) {
	pmts, err := p.dao.FindExpiredPayment(ctx, offset, limit, t)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Payment, 0, len(pmts))
	for _, pmt := range pmts {
		res = append(res, p.toDomain(pmt))
	}
	return res, nil
}

func NewPaymentRepository(dao dao.PaymentDAO) PaymentRepository {
	return &paymentRepository{
		dao: dao,
	}
}

func (p *paymentRepository) GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error) {
	r, err := p.dao.GetPayment(ctx, bizTradeNO)
	return p.toDomain(r), err
}

func (p *paymentRepository) AddPayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.Insert(ctx, p.toEntity(pmt))
}

func (p *paymentRepository) UpdatePayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.UpdateTxnIDAndStatus(ctx, pmt.BizTradeNO, pmt.TxnID, pmt.Status)
}

// CreateLocalMessage assignment week17
func (p *paymentRepository) CreateLocalMessage(ctx context.Context, bizTradeNO string, status domain.PaymentStatus) error {
	return p.dao.CreateLocalMessage(ctx, bizTradeNO, status)
}

func (p *paymentRepository) FindMessage(ctx context.Context, offset int, limit int) ([]domain.PaymentMessage, error) {
	l, err := p.dao.FindMessage(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	d := make([]domain.PaymentMessage, 0, len(l))
	for _, msg := range l {
		d = append(d, p.toDomainMsg(msg))
	}
	return d, nil
}

func (p *paymentRepository) UpdateMessageById(ctx context.Context, id int64) error {
	return p.dao.UpdateMessageById(ctx, id)
}

func (p *paymentRepository) toEntity(pmt domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatusInit,
	}
}

func (p *paymentRepository) toDomain(pmt dao.Payment) domain.Payment {
	return domain.Payment{
		Amt: domain.Amount{
			Currency: pmt.Currency,
			Total:    pmt.Amt,
		},
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatus(pmt.Status),
		TxnID:       pmt.TxnID.String,
	}
}

func (p *paymentRepository) toDomainMsg(pmt dao.LocalMessage) domain.PaymentMessage {
	return domain.PaymentMessage{
		Id:         pmt.Id,
		BizTradeNO: pmt.BizTradeNO,
		Status:     pmt.Status,
	}
}

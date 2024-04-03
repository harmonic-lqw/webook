package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
	"webook/payment/domain"
)

type PaymentGORMDAO struct {
	db *gorm.DB
}

func (p *PaymentGORMDAO) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error) {
	var res []Payment
	err := p.db.WithContext(ctx).
		Where("status = ? AND utime < ?", uint8(domain.PaymentStatusInit), t.UnixMilli()).
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (p *PaymentGORMDAO) UpdateTxnIDAndStatus(ctx context.Context, bizTradeNO string, txnID string, status domain.PaymentStatus) error {
	return p.db.WithContext(ctx).Model(&Payment{}).
		Where("biz_trade_no = ?", bizTradeNO).
		Updates(map[string]any{
			"txn_id": txnID,
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func NewPaymentGORMDAO(db *gorm.DB) PaymentDAO {
	return &PaymentGORMDAO{
		db: db,
	}
}

func (p *PaymentGORMDAO) Insert(ctx context.Context, pmt Payment) error {
	now := time.Now().UnixMilli()
	pmt.Utime = now
	pmt.Ctime = now
	return p.db.WithContext(ctx).Create(&pmt).Error
}

func (p *PaymentGORMDAO) GetPayment(ctx context.Context, bizTradeNO string) (Payment, error) {
	var res Payment
	err := p.db.WithContext(ctx).Where("biz_trade_no = ?", bizTradeNO).First(&res).Error
	return res, err
}

// CreateLocalMessage assignment week17
func (p *PaymentGORMDAO) CreateLocalMessage(ctx context.Context, bizTradeNO string, status domain.PaymentStatus) error {
	return p.db.WithContext(ctx).Create(&LocalMessage{
		BizTradeNO: bizTradeNO,
		Status:     status.AsUint8(),
		Send:       messageNotSend,
	}).Error
}

func (p *PaymentGORMDAO) FindMessage(ctx context.Context, offset int, limit int) ([]LocalMessage, error) {
	var res []LocalMessage
	err := p.db.WithContext(ctx).Where("send = ?", messageNotSend).
		Offset(offset).
		Limit(limit).Find(&res).Error
	return res, err
}

func (p *PaymentGORMDAO) UpdateMessageById(ctx context.Context, id int64) error {
	return p.db.WithContext(ctx).Where("id = ?", id).Updates(map[string]any{
		"send": messageSend,
	}).Error
}

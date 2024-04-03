package dao

import (
	"context"
	"database/sql"
	"time"
	"webook/payment/domain"
)

type PaymentDAO interface {
	Insert(ctx context.Context, pmt Payment) error
	GetPayment(ctx context.Context, bizTradeNO string) (Payment, error)
	UpdateTxnIDAndStatus(ctx context.Context, bizTradeNO string, txnID string, status domain.PaymentStatus) error
	FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error)

	// CreateLocalMessage assignment week17
	CreateLocalMessage(ctx context.Context, bizTradeNO string, status domain.PaymentStatus) error
	FindMessage(ctx context.Context, offset int, limit int) ([]LocalMessage, error)
	UpdateMessageById(ctx context.Context, id int64) error
}

type Payment struct {
	Id          int64 `gorm:"primaryKey, autoIncrement" bson:"id, omitempty"`
	Amt         int64
	Currency    string
	Description string `gorm:"description"`

	// 业务方传过来的
	BizTradeNO string `gorm:"column:biz_trade_no;type:varchar(256);unique"`

	// 第三方支付平台的事务，唯一 ID
	TxnID sql.NullString `gorm:"column:txn_id;type:varchar(128);unique"`

	Status uint8
	Utime  int64
	Ctime  int64
}

type LocalMessage struct {
	Id         int64  `gorm:"primaryKey, autoIncrement"`
	BizTradeNO string `gorm:"column:biz_trade_no;index"`
	Status     uint8
	Send       int `gorm:"index"`
}

const (
	messageSend = iota
	messageNotSend
)

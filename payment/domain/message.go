package domain

type PaymentMessage struct {
	Id         int64
	BizTradeNO string
	Status     uint8
}

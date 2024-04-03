package domain

type Amount struct {
	// 货币类型
	Currency string
	// 余额，精确到分，但不同的货币也不同
	Total int64
}

type Payment struct {
	Amt Amount

	BizTradeNO  string
	Description string
	Status      PaymentStatus
	TxnID       string
}

type PaymentStatus uint8

func (s PaymentStatus) AsUint8() uint8 {
	return uint8(s)
}

const (
	PaymentStatusUnknown = iota
	PaymentStatusInit    = iota
	PaymentStatusSuccess = iota
	PaymentStatusFailed  = iota
	PaymentStatusRefund  = iota
)

package wechat

import (
	"context"
	"errors"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"time"
	"webook/payment/domain"
	"webook/payment/events"
	"webook/payment/repository"
	"webook/pkg/logger"
)

var errUnknownTransactionState = errors.New("未知的微信事务状态")

type NativePaymentService struct {
	svc       *native.NativeApiService
	appID     string
	mchID     string
	notifyURL string
	repo      repository.PaymentRepository
	l         logger.LoggerV1
	producer  events.Producer

	// 在微信 native 里面，分别是
	// SUCCESS：支付成功
	// REFUND：转入退款
	// NOTPAY：未支付
	// CLOSED：已关闭
	// REVOKED：已撤销（付款码支付）
	// USERPAYING：用户支付中（付款码支付）
	// PAYERROR：支付失败(其他原因，如银行返回失败)
	// 因此这里需要映射到我们内部的订单状态
	nativeCBTypeToStatus map[string]domain.PaymentStatus
}

func NewNativePaymentService(svc *native.NativeApiService,
	repo repository.PaymentRepository,
	l logger.LoggerV1,
	producer events.Producer,
	appid, mchid string) *NativePaymentService {
	return &NativePaymentService{
		producer:  producer,
		l:         l,
		repo:      repo,
		svc:       svc,
		appID:     appid,
		mchID:     mchid,
		notifyURL: "https://localhost:8086/pay/callback",
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"CLOSED":   domain.PaymentStatusFailed,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
		},
	}
}

func (n *NativePaymentService) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
	err := n.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}
	resp, _, err := n.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(n.appID),
		Mchid:       core.String(n.mchID),
		Description: core.String(pmt.Description),
		OutTradeNo:  core.String(pmt.BizTradeNO),
		TimeExpire:  core.Time(time.Now().Add(time.Minute * 30)),
		NotifyUrl:   core.String(n.notifyURL),
		Amount: &native.Amount{
			Currency: core.String(pmt.Amt.Currency),
			Total:    core.Int64(pmt.Amt.Total),
		},
	})
	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil
}

func (n *NativePaymentService) GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error) {
	return n.repo.GetPayment(ctx, bizTradeId)
}

func (n *NativePaymentService) HandleCallback(ctx context.Context, transaction *payments.Transaction) error {
	return n.updateByTxn(ctx, transaction)
}

func (n *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	status, ok := n.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, %s", errUnknownTransactionState, *txn.TradeState)
	}
	pmt := domain.Payment{
		BizTradeNO: *txn.OutTradeNo,
		TxnID:      *txn.TransactionId,
		Status:     status,
	}
	err := n.repo.UpdatePayment(ctx, pmt)
	if err != nil {
		// 这里可能有很多问题，有可能数据库更新成功，也有可能更新失败，因此需要依靠定时对账来保证数据库的正确
		return err
	}

	err1 := n.producer.ProducePaymentEvent(ctx, events.PaymentEvent{
		BizTradeNO: pmt.BizTradeNO,
		Status:     pmt.Status.AsUint8(),
	})

	if err1 != nil {
		// 一定要做好监控和告警
		//n.l.Error("发送支付事件失败", logger.Error(err),
		//	logger.String("biz_trade_no", pmt.BizTradeNO))

		// assignment week17
		ctxMsg, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = n.repo.CreateLocalMessage(ctxMsg, pmt.BizTradeNO, pmt.Status)
		if err != nil {
			n.l.Error("存储失败消息到数据库失败", logger.Error(err),
				logger.String("biz_trade_no", pmt.BizTradeNO))
		}
	}
	// 这里可以返回 nil 因为数据库成功更新了
	return nil
}

func (n *NativePaymentService) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayment(ctx, offset, limit, t)
}

func (n *NativePaymentService) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
	txn, _, err := n.svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(bizTradeNO),
		Mchid:      core.String(n.mchID),
	})
	if err != nil {
		return err
	}
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) FindMessage(ctx context.Context, offset int, limit int) ([]domain.PaymentMessage, error) {
	return n.repo.FindMessage(ctx, offset, limit)
}

func (n *NativePaymentService) UpdateMessageById(ctx context.Context, id int64) error {
	return n.repo.UpdateMessageById(ctx, id)
}

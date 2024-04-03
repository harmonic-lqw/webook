package ioc

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"webook/payment/events"
	"webook/payment/repository"
	"webook/payment/service/wechat"
	"webook/pkg/logger"
)

func InitWechatClient(cfg WechatConfig) *core.Client {
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(
		cfg.KeyPath,
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client, err := core.NewClient(
		ctx,
		option.WithWechatPayAutoAuthCipher(
			cfg.MchID, cfg.MchSerialNum,
			mchPrivateKey, cfg.MchKey),
	)
	if err != nil {
		panic(err)
	}
	return client
}

func InitWechatNativeService(client *core.Client,
	repo repository.PaymentRepository,
	l logger.LoggerV1,
	producer events.Producer,
	cfg WechatConfig) *wechat.NativePaymentService {
	return wechat.NewNativePaymentService(&native.NativeApiService{
		Client: client,
	}, repo, l, producer, cfg.AppID, cfg.MchID)
}

func InitWechatConfig() WechatConfig {
	// 都没有 T_T...
	return WechatConfig{
		AppID:        "",
		MchID:        "",
		MchKey:       "",
		MchSerialNum: "",
		CertPath:     "",
		KeyPath:      "",
	}
}

type WechatConfig struct {
	AppID        string
	MchID        string
	MchKey       string
	MchSerialNum string

	// 证书
	CertPath string
	KeyPath  string
}

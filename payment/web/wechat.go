package web

import (
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"net/http"
	"webook/payment/service/wechat"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type WechatHandler struct {
	handler   *notify.Handler
	l         logger.LoggerV1
	nativeSvc *wechat.NativePaymentService
}

func NewWechatHandler(handler *notify.Handler, nativeSvc *wechat.NativePaymentService, l logger.LoggerV1) *WechatHandler {
	return &WechatHandler{
		handler:   handler,
		nativeSvc: nativeSvc,
		l:         l,
	}
}

func (h *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.POST("/pay/callback", ginx.Wrap(h.HandleNative))
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) (ginx.Result, error) {
	transaction := &payments.Transaction{}
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		return ginx.Result{}, err
	}
	err = h.nativeSvc.HandleCallback(ctx, transaction)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "处理回调失败，你那边重新发送吧",
		}, err
	}
	return ginx.Result{}, nil
}

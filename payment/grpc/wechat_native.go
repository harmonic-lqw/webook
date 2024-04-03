package grpc

import (
	"context"
	"google.golang.org/grpc"
	pmtv1 "webook/api/proto/gen/payment/v1"
	"webook/payment/domain"
	"webook/payment/service/wechat"
)

type WechatServiceServer struct {
	pmtv1.UnimplementedWechatPaymentServiceServer
	svc *wechat.NativePaymentService
}

func NewWechatServiceServer(svc *wechat.NativePaymentService) *WechatServiceServer {
	return &WechatServiceServer{svc: svc}
}

func (s *WechatServiceServer) Register(server *grpc.Server) {
	pmtv1.RegisterWechatPaymentServiceServer(server, s)
}

func (s *WechatServiceServer) NativePrePay(ctx context.Context, request *pmtv1.PrePayRequest) (*pmtv1.NativePrePayResponse, error) {
	codeURL, err := s.svc.Prepay(ctx, domain.Payment{
		Amt: domain.Amount{
			Currency: request.Amt.Currency,
			Total:    request.Amt.Total,
		},
		BizTradeNO:  request.BizTradeNo,
		Description: request.Description,
	})
	if err != nil {
		return nil, err
	}
	return &pmtv1.NativePrePayResponse{
		CodeUrl: codeURL,
	}, nil
}

func (s *WechatServiceServer) GetPayment(ctx context.Context, req *pmtv1.GetPaymentRequest) (*pmtv1.GetPaymentResponse, error) {
	p, err := s.svc.GetPayment(ctx, req.GetBizTradeNo())
	if err != nil {
		return nil, err
	}
	return &pmtv1.GetPaymentResponse{
		Status: pmtv1.PaymentStatus(p.Status),
	}, nil
}

package auth

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
	key []byte
}

func NewSMSService(svc sms.Service, key []byte) *SMSService {
	return &SMSService{
		svc: svc,
		key: key,
	}
}

func (s *SMSService) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims SMSClaims
	// 如果 jwt 解析通过(err == nil)，就认为通过了身份验证
	_, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	return s.svc.Send(ctx, claims.tplId, args, numbers...)
}

type SMSClaims struct {
	jwt.RegisteredClaims
	tplId string
	// 还可以额外加字段 比如身份认证的一些额外辅助信息
}

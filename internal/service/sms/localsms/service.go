package localsms

import (
	"context"
	"log"
)

type Service struct {
}

func (s Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	log.Println("验证码： ", args)
	return nil
}

func NewService() *Service {
	return &Service{}
}

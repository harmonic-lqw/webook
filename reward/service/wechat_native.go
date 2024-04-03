package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/bloom"
	"strconv"
	"strings"
	accountv1 "webook/api/proto/gen/account/v1"
	pmtv1 "webook/api/proto/gen/payment/v1"
	"webook/pkg/logger"
	"webook/reward/domain"
	"webook/reward/repository"
)

type WechatNativeRewardService struct {
	client  pmtv1.WechatPaymentServiceClient
	repo    repository.RewardRepository
	l       logger.LoggerV1
	aClient accountv1.AccountServiceClient

	// assignment week17 使用布隆过滤器
	bloomFilter *bloom.Filter
}

func NewWechatNativeRewardService(client pmtv1.WechatPaymentServiceClient, repo repository.RewardRepository,
	l logger.LoggerV1, aClient accountv1.AccountServiceClient) RewardService {
	return &WechatNativeRewardService{client: client, repo: repo, l: l, aClient: aClient}
}

func (s *WechatNativeRewardService) PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	codeUrl, err := s.repo.GetCachedCodeURL(ctx, r)
	if err == nil {
		return codeUrl, nil
	}
	r.Status = domain.RewardStatusInit
	rid, err := s.repo.CreateReward(ctx, r)
	if err != nil {
		return domain.CodeURL{}, err
	}
	repo, err := s.client.NativePrePay(ctx, &pmtv1.PrePayRequest{
		Amt: &pmtv1.Amount{
			Total:    r.Amt,
			Currency: "CNY",
		},
		BizTradeNo:  fmt.Sprintf("reward-%d", rid),
		Description: fmt.Sprintf("打赏-%s", r.Target.BizName),
	})

	if err != nil {
		return domain.CodeURL{}, err
	}

	cu := domain.CodeURL{
		Rid: rid,
		URL: repo.CodeUrl,
	}

	err1 := s.repo.CachedCodeURL(ctx, cu, r)
	if err1 != nil {
		s.l.Error("缓存二维码失败", logger.Error(err1))
	}
	return cu, err
}

func (s *WechatNativeRewardService) GetReward(ctx context.Context, rid int64, uid int64) (domain.Reward, error) {
	r, err := s.repo.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	if r.SrcUid != uid {
		return domain.Reward{}, errors.New("不是该打赏人创建的打赏")
	}
	if r.Completed() {
		return r, nil
	}

	// 这里是一条慢路径，去支付服务那边查看支付状态（支付那边也是一路去数据库中查）
	resp, err := s.client.GetPayment(ctx, &pmtv1.GetPaymentRequest{
		BizTradeNo: s.bizTradeNO(r.Id),
	})

	if err != nil {
		// 查支付失败，直接返回这边数据库查询的结果
		s.l.Error("获得打赏结果慢路径（去支付服务那边查询）执行失败",
			logger.Int64("rid", r.Id), logger.Error(err))
		return r, nil
	}

	// 根据支付状态更新打赏状态
	switch resp.Status {
	case pmtv1.PaymentStatus_PaymentStatusFailed:
		r.Status = domain.RewardStatusFailed
	case pmtv1.PaymentStatus_PaymentStatusSuccess:
		r.Status = domain.RewardStatusPayed
	case pmtv1.PaymentStatus_PaymentStatusInit:
		r.Status = domain.RewardStatusInit
	case pmtv1.PaymentStatus_PaymentStatusRefund:
		r.Status = domain.RewardStatusFailed
	}

	err = s.repo.UpdateStatus(ctx, rid, r.Status)
	if err != nil {
		s.l.Error("更新打赏状态失败", logger.Int64("rid", r.Id), logger.Error(err))
		return r, nil
	}
	return r, nil
}

// UpdateReward 收到支付那边的成功支付的消息通知
func (s *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error {
	rid := s.toRid(bizTradeNO)
	err := s.repo.UpdateStatus(ctx, rid, status)
	if err != nil {
		return err
	}

	// 完成支付，准备入账
	if status == domain.RewardStatusPayed {
		r, err := s.repo.GetReward(ctx, rid)
		if err != nil {
			return err
		}

		// 布隆过滤器
		bloomKey := s.getBloomKey(r.Target.Biz, r.Target.BizId)
		ok, er := s.bloomFilter.Exists([]byte(bloomKey))
		if er != nil {
			return er
		}
		if ok {
			// 如果存在，表示已经记过账，直接返回
			// 这里可能有假阳性问题
			return nil
		}

		weAmt := int64(float64(r.Amt) * 0.1)
		_, err = s.aClient.Credit(ctx, &accountv1.CreditRequest{
			Biz:   "reward",
			BizId: rid,
			Items: []*accountv1.CreditItem{
				{
					AccountType: accountv1.AccountType_AccountTypeReward,
					Amt:         weAmt,
					Currency:    "CNY",
				},
				{
					Account:     r.SrcUid,
					Uid:         r.SrcUid,
					AccountType: accountv1.AccountType_AccountTypeReward,
					Amt:         r.Amt - weAmt,
					Currency:    "CNY",
				},
			},
		})
		if err != nil {
			s.l.Error("入账失败，请修数据！！！",
				logger.String("biz_trade_no", bizTradeNO),
				logger.Error(err))
			return err
		}

		// 对账成功，写入布隆过滤器
		er = s.bloomFilter.Add([]byte(bloomKey))
		if er != nil {
			s.l.Error("入账成功但写入布隆过滤器失败",
				logger.Error(er),
				logger.String("bloom_key", bloomKey))
		}

	}
	return nil
}

func (s *WechatNativeRewardService) bizTradeNO(rId int64) string {
	return fmt.Sprintf("reward-%d", rId)
}

func (s *WechatNativeRewardService) toRid(bizTradeNO string) int64 {
	ridStr := strings.Split(bizTradeNO, "-")
	val, _ := strconv.ParseInt(ridStr[1], 10, 64)
	return val
}

func (s *WechatNativeRewardService) getBloomKey(biz string, bizId int64) string {
	return fmt.Sprintf("%s+%d", biz, bizId)
}

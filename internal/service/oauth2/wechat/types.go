package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"webook/internal/domain"
	"webook/pkg/logger"
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

// ? 暂时无法注册
var redirectURL = url.PathEscape(``)

type service struct {
	appID     string
	appSecret string
	client    *http.Client
	l         logger.LoggerV1
}

func NewService(appID string, appSecret string, l logger.LoggerV1) Service {
	return &service{
		appID:     appID,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	const authURLPattern = `https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect`
	return fmt.Sprintf(authURLPattern, s.appID, redirectURL, state), nil
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	accessTokenURL := fmt.Sprintf(`https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code`,
		s.appID, s.appSecret, code)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenURL, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	var res Result
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		// 序列化出错
		return domain.WechatInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("微信接口调用失败 errcode: %s, errmsg: %s", res.ErrCode, res.ErrMsg)
	}
	return domain.WechatInfo{
		UnionId: res.UnionId,
		OpenId:  res.OpenId,
	}, nil

}

type Result struct {
	// 接口调用凭证
	// 你后面就可以拿着这个 access_token 去访问微信，获取用户的数据
	AccessToken string `json:"access_token"`
	// access_token接口调用凭证超时时间，单位（秒）
	// 也就是 access_token 的有效期
	ExpiresIn int64 `json:"expires_in"`
	// 用户刷新 access_token
	// 当 access_token 过期之后，你就可以用这个 refresh_token 去找微信换一个新的access_token
	RefreshToken string `json:"refresh_token"`
	// 授权用户唯一标识
	// 你在这个应用下的唯一 ID
	OpenId string `json:"openid"`
	// 用户授权的作用域，使用逗号（,）分隔
	Scope string `json:"scope"`
	// 当且仅当该网站应用已获得该用户的userinfo授权时，才会出现该字段
	// 你在这个公司下的唯一 ID
	UnionId string `json:"unionid"`

	// 错误返回

	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

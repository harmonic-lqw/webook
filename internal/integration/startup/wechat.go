package startup

import (
	"webook/internal/service/oauth2/wechat"
	"webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	appID := ""
	appSecret := ""
	return wechat.NewService(appID, appSecret, l)
}

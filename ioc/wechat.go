package ioc

import (
	"webook/internal/service/oauth2/wechat"
	"webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	//appID, ok := os.LookupEnv("WECHAT_APP_ID")
	//if !ok {
	//	panic("找不到环境变量 WECHAT_APP_ID")
	//}
	//appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	//if !ok {
	//	panic("找不到环境变量 WECHAT_APP_SECRET")
	//}
	appID := "wx7256bc69ab349c72" // 大明老师 appID
	appSecret := ""               // ? 暂时无法注册

	return wechat.NewService(appID, appSecret, l)
}

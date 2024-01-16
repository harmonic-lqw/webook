package startup

import (
	"webook/internal/service/sms"
	"webook/internal/service/sms/localsms"
)

func InitSMSService() sms.Service {
	return initMemorySMSService()
}

func initMemorySMSService() sms.Service {
	return localsms.NewService()
}

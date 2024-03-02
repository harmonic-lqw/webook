package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"time"
	"webook/internal/service/sms"
	"webook/internal/service/sms/auth"
	"webook/internal/service/sms/failover"
	"webook/internal/service/sms/localsms"
	"webook/internal/service/sms/ratelimit"
	"webook/internal/service/sms/tencent"
	"webook/pkg/limiter"
)

// InitSMSServiceV1 使用装饰器进行初始化
func InitSMSServiceV1(redisClient redis.Cmdable) sms.Service {
	svc := initTencentSMSService()
	svc = ratelimit.NewRateLimitSMSService(svc, limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000))
	svc = failover.NewFailOverSMSService([]sms.Service{svc})
	svc = auth.NewSMSService(svc, []byte("this is the JWT TOKEN given by the SMS platform"))
	return svc
}

func InitSMSService() sms.Service {
	return initTencentSMSService()
}

func initTencentSMSService() sms.Service {
	// 在这里你也可以考虑从配置文件里面读取
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("没有找到环境变量 SMS_SECRET_ID ")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("没有找到环境变量 SMS_SECRET_KEY")
	}

	c, err := tencentSMS.NewClient(common.NewCredential(secretId, secretKey),
		"ap-nanjing",
		profile.NewClientProfile())
	if err != nil {
		panic(err)
	}

	return tencent.NewService(c, "1400877785", "泛古玉的个人公众号")
}

func InitMemorySMSService() sms.Service {
	return localsms.NewService()
}

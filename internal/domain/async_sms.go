package domain

type AsyncSms struct {
	Id      int64
	TplId   string
	Args    []string
	Numbers []string
	// 重试的最大次数
	RetryMax int
}

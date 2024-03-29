package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Password string

	Phone string

	// UTC 0 的时区
	Ctime time.Time

	// 编辑字段
	NickName string
	Birthday string
	AboutMe  string

	WechatInfo WechatInfo
}

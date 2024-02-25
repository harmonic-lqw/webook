package ginx

import "github.com/golang-jwt/jwt/v5"

// Result 你可以通过在 Result 里面定义更加多的字段，来配合 Wrap 方法
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId    int64
	Ssid      string
	UserAgent string
}

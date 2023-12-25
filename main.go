package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"webook/internal/web/middleware"
)

func main() {
	server := InitWebServer()

	//server.GET("/hello", func(ctx *gin.Context) {
	//	//	ctx.String(http.StatusOK, "成功部署")
	//	//})

	server.Run(":8080")
}

func useSession(server *gin.Engine) {
	// 基于 cookie 的实现
	// 存储数据的，也就是 userId 存的地方
	// 数据直接存储在 cookie 里
	store := cookie.NewStore([]byte("secret"))

	// 基于内存的实现
	// 第一个参数是 authentication key，第二个是 encryption key
	// Authentication：是指身份认证。
	// Encryption：是指数据加密。
	//这两者再加上授权（权限控制），就是信息安全的三个核心概念。
	//store := memstore.NewStore([]byte("ezFwyB49AsChLgJIlqo2BxeUEOtS80tyorXuTt78bXslhPRZxTAlGIzUZWq0lU7X"),
	//	[]byte("kiRmG9M1YWzOojPNhIXtmrP52nPA1IfFCf5UG6TweEYhV79UdBKgzVbMRaLNAdug"))

	// 基于 Redis 的实现
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	//	[]byte("pBnSDaa0FCypBlPSpSoATWB4VZIS9niS"),
	//	[]byte("pBnSDaa0FCypBlPSpSoATWB4VZIS9niA"))
	//if err != nil {
	//	panic(err)
	//}

	// 初始化 session
	server.Use(sessions.Sessions("ssid", store))

	// 登录校验
	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}

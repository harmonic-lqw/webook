package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"time"
	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/internal/web/middleware"
	"webook/pkg/ginx/middleware/ratelimit"

	"github.com/redis/go-redis/v9"
)

func main() {
	//db := initDB()
	//
	//server := initWebServer()
	//
	//initUserHdl(db, server)

	server := gin.Default()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "成功部署")
	})

	server.Run(":8080")
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)

	hdl := web.NewUserHandler(us)
	hdl.RegisterRoutes(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:123456@tcp(127.0.0.1:13316)/webook"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(
		cors.New(cors.Config{
			//AllowAllOrigins: true,
			//AllowOrigins:     []string{"http://localhost:3000"},
			AllowCredentials: true,
			AllowHeaders:     []string{"authorization", "content-type"},

			ExposeHeaders: []string{"x-jwt-token"}, // 允许前端访问后端响应中带的头部
			AllowOriginFunc: func(origin string) bool {
				return true
			},
			MaxAge: 12 * time.Hour,
		}),
	)

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	useJWT(server)
	//useSession(server)

	return server
}

func useJWT(server *gin.Engine) {
	// 登录校验
	login := &middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())
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

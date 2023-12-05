package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		// 不需要登录校验
		if path == "/users/signup" || path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			// 中止：表示没有登录，不要往后执行、也不要执行后面的业务逻辑
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()
		const updateTimeKey = "update_time"
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		// updateTimeKey未设置、断言失败、超时，都将触发同样的session刷新
		if val == nil || (!ok) || (now.Sub(lastUpdateTime) > time.Minute*5) {
			sess.Set(updateTimeKey, now)
			sess.Set("userId", userId)
			sess.Options(sessions.Options{
				// 十五分钟
				MaxAge: 900,
			})
			err := sess.Save()
			if err != nil {
				fmt.Println(err)
			}
		}

	}
}

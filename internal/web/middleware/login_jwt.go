package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	ijwt "webook/internal/web/jwt"
)

type LoginJWTMiddlewareBuilder struct {
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(hdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: hdl,
	}
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		// 不需要登录校验
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms" ||
			path == "/users/login_sms/code/send" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			return
		}
		tokenStr := m.ExtractToken(ctx)
		var uc ijwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.JWTKey, nil
		})
		if err != nil {
			// token 无法解析：token 不对，token 是伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			// token 解析成功：token 非法或过期
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 压测 profile 要关闭
		// 解析出来的 User-Agent（它一定是登录时放入的，因为如果 jwt-token 被修改，是不会解析成功的） != 该次请求携带的 User-Agent
		if uc.UserAgent != ctx.GetHeader("User-Agent") {
			// 只要进来这个分支，大概率是攻击者，也有可能是浏览器升级等
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = m.CheckSession(ctx, uc.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 设置了长短 token 后不再需要这些定时刷新机制
		//expireTime := uc.ExpiresAt
		//if expireTime.Before(time.Now()) {
		//	// token 过期
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		//
		//// 过期时间设置 30 分钟，每 10 分钟刷新一次：当剩余过期时间小于 20 分钟时就应该刷新了
		//if expireTime.Sub(time.Now()) < time.Minute*20 {
		//	uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
		//	tokenStr, err = token.SignedString(web.JWTKey)
		//	ctx.Header("x-jwt-token", tokenStr)
		//	if err != nil {
		//		// 不要 panic 掉，因为仅仅是过期时间没有成功刷新，用户仍然处于登录状态，不影响使用
		//		log.Println(err)
		//	}
		//}

		ctx.Set("user", uc)

	}
}

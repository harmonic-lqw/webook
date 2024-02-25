package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/pkg/ginx"
)

type RedisJWTHandler struct {
	cmd           redis.Cmdable
	signingMethod jwt.SigningMethod
	rcExpiration  time.Duration
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		cmd:           cmd,
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  time.Hour * 24 * 7,
	}
}

var JWTKey = []byte("pBnSDaa0oCypBlPSpSoATWB4VZIS9niS")
var RCJWTKey = []byte("pBnSDaa0oCypBlPSpSoATWB4VZIS9niB")

// ExtractToken 根据约定，token 在 Authorization 头部
// Bearer XXX
func (rh *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return ""
	}

	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		return ""
	}
	tokenStr := segs[1]
	return tokenStr
}

func (rh *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims)
	// 长 token 过期时间
	return rh.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid), "", rh.rcExpiration).Err()
}

func (rh *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := rh.SetRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return rh.SetJWTToken(ctx, uid, ssid)
}

func (rh *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		UserId:    uid,
		Ssid:      ssid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			// 30分钟后过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(rh.signingMethod, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Set("user", uc)
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (rh *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		UserId: uid,
		Ssid:   ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			// 长 token 设置 7 天后过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(rh.rcExpiration)),
		},
	}
	token := jwt.NewWithClaims(rh.signingMethod, rc)
	tokenStr, err := token.SignedString(RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (rh *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	// 是否已经退出登录的校验
	cnt, err := rh.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err == nil && cnt > 0 {
		// cnt > 0 说明这个 token 在 redis 中，意味它是无效的
		// || 代表一种降级策略，err == nil 表示 redis 没问题，那么就执行严格的 ssid 校验，如果 redis 出了问题，err != nil，就不需要严格校验 ssid
		return errors.New("token 无效")
	}
	return nil

	// 严格的方式，redis 一旦出问题就不再能够通过登录校验
	//cnt, err := rh.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	//if err != nil {
	//	return err
	//}
	//if cnt > 0 {
	//	return errors.New("token 无效")
	//}
	//return nil

}

type UserClaims = ginx.UserClaims

type RefreshClaims struct {
	jwt.RegisteredClaims
	UserId int64
	Ssid   string
}

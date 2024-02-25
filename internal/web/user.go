// Package web
// Representing the HTTP interface, a package that directly interacts with the HTTP interface
package web

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"webook/internal/domain"
	"webook/internal/errs"
	"webook/internal/service"
	ijwt "webook/internal/web/jwt"
	"webook/pkg/ginx"
)

const (
	emailRegexPattern    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	maxAboutMeRune       = 200
	maxNickNameRune      = 10
	bizLogin             = "login"
)

type UserHandler struct {
	ijwt.Handler
	emailRegExp    *regexp.Regexp
	passwordRegExp *regexp.Regexp
	svc            service.UserService
	codeSvc        service.CodeService

	//l logger.LoggerV1
}

func NewUserHandler(svc service.UserService,
	hdl ijwt.Handler,
	codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		emailRegExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
		codeSvc:        codeSvc,
		Handler:        hdl,

		//l: l,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	//server.POST("/users/login", h.Login)
	//server.POST("/users/signup", h.SignUp)
	//server.POST("/users/edit", h.Edit)
	//server.GET("/users/profile", h.Profile)

	ug := server.Group("/users")
	//ug.POST("/login", h.Login)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/logout", h.LogoutJWT)
	ug.POST("/signup", ginx.WrapReq[SignUpReq](h.SignUp))
	//ug.POST("/edit", h.Edit)
	ug.POST("/edit", h.EditJWT)
	//ug.GET("/profile", h.Profile)
	ug.GET("/profile", h.ProfileJWT)
	ug.GET("/refresh_token", h.RefreshToken)

	// 手机验证码登录相关
	ug.POST("/login_sms/code/send", h.SendSMSLoginCode)
	ug.POST("/login_sms", h.LoginSMS)

}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.Error(err)
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "请输入手机号码",
		})
	}
	err := h.codeSvc.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		// 打印日志 Warn
		// 少数错误可以接受，频繁出现需要注意时，考虑使用 Warning
		//zap.L().Warn("频繁发送验证码")
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	// errors.New("验证码存在，但是没有过期时间") 也会进来这个分支
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}

}

func (h *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.Error(err)
		return
	}
	ok, err := h.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统异常",
		})
		// 打印日志 Error
		// 比较严重的错误，基本上需要立即去处理
		// 注意敏感信息，生产环境不可以打印敏感信息，开发环境可以
		//zap.L().Error("手机验证码验证失败", zap.Error(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码不对，请重新输入",
		})
		return
	}
	// 因为需求：手机号第一次登录需要自动注册，所以需要一个新的方法
	u, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = h.SetLoginToken(ctx, u.Id)
	if err != nil {
		ctx.Error(err)
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	u, err := h.svc.Login(ctx, req.Email, req.Password)

	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 十五分钟
			MaxAge: 900,
		})
		err = sess.Save()
		if err != nil {
			ctx.Error(err)
			ctx.String(http.StatusOK, "服务器异常")
			return
		}

		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户不存在或密码错误")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	u, err := h.svc.Login(ctx, req.Email, req.Password)

	switch err {
	case nil:
		err = h.SetLoginToken(ctx, u.Id)
		if err != nil {
			ctx.Error(err)
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.Error(err)
		ctx.String(http.StatusOK, "用户不存在或密码错误")
	default:
		ctx.Error(err)
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.Error(err)
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "退出登录成功",
	})
	return

}

func (h *UserHandler) ProfileJWT(ctx *gin.Context) {
	type UserInfoResp struct {
		Nickname string
		Email    string
		Phone    string
		Birthday string
		AboutMe  string
	}
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	u, err := h.svc.GetUserInfo(ctx, uc.UserId)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, UserInfoResp{
			Nickname: u.NickName,
			Email:    u.Email,
			Birthday: u.Birthday,
			AboutMe:  u.AboutMe,
		})
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	type UserInfoResp struct {
		Nickname string
		Email    string
		Phone    string
		Birthday string
		AboutMe  string
	}
	sess := sessions.Default(ctx)
	userId := sess.Get("userId").(int64)
	u, err := h.svc.GetUserInfo(ctx, userId)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, UserInfoResp{
			Nickname: u.NickName,
			Email:    u.Email,
			Birthday: u.Birthday,
			AboutMe:  u.AboutMe,
		})
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (h *UserHandler) SignUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {
	// 使用 wrap 后，不需要再 Bind 请求了
	//if err := ctx.Bind(&req); err != nil {
	//	ctx.String(http.StatusOK, "系统错误")
	//	return Result{
	//		Code: errs.UserInternalServerError,
	//		Msg:  "系统错误",
	//	}, err
	//}

	isEmail, err := h.emailRegExp.MatchString(req.Email)

	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误，邮箱匹配超时",
		}, err
	}
	if !isEmail {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "非法邮箱格式",
		}, err
	}

	isPassword, err := h.passwordRegExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误：密码匹配超时",
		}, err
	}
	if !isPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "密码格式不对：密码必须包含字母、数字、特殊字符，并且长度不能小于 8 位",
		}, err
	}

	if req.Password != req.ConfirmPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "两次输入密码不同",
		}, err
	}

	err = h.svc.SignUp(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		return ginx.Result{
			Msg: "注册成功",
		}, nil
	case service.ErrDuplicateUser:
		return ginx.Result{
			Code: errs.UserDuplicateEmail,
			Msg:  "该邮箱已被注册",
		}, err
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "该邮箱已被注册",
		}, err
	}

}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		NickName string `json:"nickname"`
		Birthday string `json:"birthday"` // 对生日的格式化(string2time)和反格式化在repository进行，但建议在web这一层
		AboutMe  string `json:"aboutMe"`
	}

	var req EditReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	// 输入检查
	// 校验生日
	birthdayParts := strings.Split(req.Birthday, "-")
	req.Birthday = fmt.Sprintf("%04s-%02s-%02s", birthdayParts[0], birthdayParts[1], birthdayParts[2])

	// 其实这种检验在前端进行即可
	// 虽然考虑到有些攻击者绕开前端，但是数据库层面是有对数据长度的限制的，如果超出长度，会返回系统错误，而对于攻击者，无需考虑其用户体验（对该错误进行判断和提示）
	if len([]rune(req.NickName)) >= maxNickNameRune {
		ctx.String(http.StatusOK, "昵称过长，请保持在10个字符（包括英文）以内")
		return
	}
	if len([]rune(req.AboutMe)) >= maxAboutMeRune {
		ctx.String(http.StatusOK, "简介过长，请保持在200个字符（包括英文）以内")
		return
	}

	sess := sessions.Default(ctx)
	userId := sess.Get("userId").(int64)
	err = h.svc.EditUserInfo(ctx, userId, req.NickName, req.Birthday, req.AboutMe)
	switch err {
	case nil:
		ctx.String(http.StatusOK, "提交成功")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) EditJWT(ctx *gin.Context) {
	type EditReq struct {
		NickName string `json:"nickname"`
		Birthday string `json:"birthday"` // 对生日的格式化(string2time)和反格式化在repository进行，但建议在web这一层
		AboutMe  string `json:"aboutMe"`
	}

	var req EditReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	// 输入检查
	// 校验生日
	birthdayParts := strings.Split(req.Birthday, "-")
	req.Birthday = fmt.Sprintf("%04s-%02s-%02s", birthdayParts[0], birthdayParts[1], birthdayParts[2])

	// 其实这种检验在前端进行即可
	// 虽然考虑到有些攻击者绕开前端，但是数据库层面是有对数据长度的限制的，如果超出长度，会返回系统错误，而对于攻击者，无需考虑其用户体验（对该错误进行判断和提示）
	if len([]rune(req.NickName)) >= maxNickNameRune {
		ctx.String(http.StatusOK, "昵称过长，请保持在10个字符（包括英文）以内")
		return
	}
	if len([]rune(req.AboutMe)) >= maxAboutMeRune {
		ctx.String(http.StatusOK, "简介过长，请保持在200个字符（包括英文）以内")
		return
	}

	uc := ctx.MustGet("user").(ijwt.UserClaims)
	err = h.svc.EditUserInfo(ctx, uc.UserId, req.NickName, req.Birthday, req.AboutMe)
	switch err {
	case nil:
		ctx.String(http.StatusOK, "提交成功")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	tokenStr := h.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RCJWTKey, nil
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 是否已经退出登录的校验
	err = h.CheckSession(ctx, rc.Ssid)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.SetJWTToken(ctx, rc.UserId, rc.Ssid)

	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

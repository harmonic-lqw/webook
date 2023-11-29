// Package web
// Representing the HTTP interface, a package that directly interacts with the HTTP interface
package web

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"webook/internal/domain"
	"webook/internal/service"
)

const (
	emailRegexPattern    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	maxAboutMeRune       = 200
	maxNickNameRune      = 10
)

type UserHandler struct {
	emailRegExp    *regexp.Regexp
	passwordRegExp *regexp.Regexp
	svc            *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRegExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	//server.POST("/users/login", h.Login)
	//server.POST("/users/signup", h.SignUp)
	//server.POST("/users/edit", h.Edit)
	//server.GET("/users/profile", h.Profile)

	ug := server.Group("/users")
	ug.POST("/login", h.Login)
	ug.POST("/signup", h.SignUp)
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)

}

func (h *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	err := ctx.Bind(&req)
	if err != nil {
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
			ctx.String(http.StatusOK, "服务器异常")
		}

		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户不存在或密码错误")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) Profile(ctx *gin.Context) {
	type UserInfoResp struct {
		NickName    string
		Email       string
		PhoneNumber string
		Birthday    string
		AboutMe     string
	}
	sess := sessions.Default(ctx)
	userId := sess.Get("userId").(int64)
	u, err := h.svc.GetUserInfo(ctx, userId)
	if err != nil {
		return
	}
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, UserInfoResp{
			NickName: u.NickName,
			Email:    u.Email,
			Birthday: u.Birthday,
			AboutMe:  u.AboutMe,
		})
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		// if failed -> return 400
		return
	}

	isEmail, err := h.emailRegExp.MatchString(req.Email)

	if err != nil {
		ctx.String(http.StatusOK, "系统错误：邮箱匹配超时")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式"+req.Email)
		return
	}

	isPassword, err := h.passwordRegExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误：密码匹配超时")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码格式不对：密码必须包含字母、数字、特殊字符，并且长度不能小于 8 位")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不同")
		return
	}

	err = h.svc.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "该邮箱已被注册")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		NickName string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}

	var req EditReq
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	// 输入检查
	birthdayParts := strings.Split(req.Birthday, "-")
	req.Birthday = fmt.Sprintf("%04s-%02s-%02s", birthdayParts[0], birthdayParts[1], birthdayParts[2])
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

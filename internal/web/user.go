// Package web
// Representing the HTTP interface, a package that directly interacts with the HTTP interface
package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	emailRegexPattern    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	emailRegExp    *regexp.Regexp
	passwordRegExp *regexp.Regexp
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		emailRegExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
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

}

func (h *UserHandler) Profile(ctx *gin.Context) {

}

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (h *UserHandler) SignUp(ctx *gin.Context) {

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		// if failed -> return 400
		return
	}

	isEmail, err := h.emailRegExp.MatchString(req.Email)

	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误：邮箱匹配超时")
		return
	}
	if !isEmail {
		ctx.String(http.StatusBadRequest, "非法邮箱格式"+req.Email)
		return
	}

	isPassword, err := h.passwordRegExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误：密码匹配超时")
		return
	}
	if !isPassword {
		ctx.String(http.StatusBadRequest, "密码格式不对：密码必须包含字母、数字、特殊字符，并且长度不能小于 8 位")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusBadRequest, "两次输入密码不同")
		return
	}

	ctx.String(http.StatusOK, "signup success")

}

func (h *UserHandler) Edit(ctx *gin.Context) {

}

package web

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/internal/domain"
	"webook/internal/service"
	svcmocks "webook/internal/service/mocks"
)

// mock 生成：mockgen -source .\internal\service\user.go -destination .\internal\service\mocks\user_mock.go -package svcmocks

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		// mock
		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService)

		// 预期输入
		reqBuilder func(t *testing.T) *http.Request

		// 预期输出
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "123456a@",
				}).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
				"email": "123@qq.com",
				"password": "123456a@",
				"confirmPassword": "123456a@"
				}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			// 非 json 字符串
			name: "Bind错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
				"email": "123@qq.com",
				"password": "123456a@",
				}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusBadRequest,
			wantBody: "",
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
				"email": "123@",
				"password": "123456a@",
				"confirmPassword": "123456a@"
				}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "非法邮箱格式" + "123@",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
				"email": "123@qq.com",
				"password": "1234",
				"confirmPassword": "1234"
				}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "密码格式不对：密码必须包含字母、数字、特殊字符，并且长度不能小于 8 位",
		},
		{
			name: "两次输入密码不匹配",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
				"email": "123@qq.com",
				"password": "123456a@",
				"confirmPassword": "123456b@"
				}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "两次输入密码不同",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "123456a@",
				}).Return(errors.New("DB err"))
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
				"email": "123@qq.com",
				"password": "123456a@",
				"confirmPassword": "123456a@"
				}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "123456a@",
				}).Return(service.ErrDuplicateUser)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
				"email": "123@qq.com",
				"password": "123456a@",
				"confirmPassword": "123456a@"
				}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "该邮箱已被注册",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// mock
			userSvc, codeSvc := tc.mock(ctrl)

			// 初始化 hdl
			hdl := NewUserHandler(userSvc, codeSvc)

			// 注册路由
			server := gin.Default()
			hdl.RegisterRoutes(server)

			// 拿到 req 和 recorder
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()

			// 发起调用
			server.ServeHTTP(recorder, req)

			// 断言
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())

		})
	}

}

//// TestEmailPattern 测试邮箱正则表达式是否正确
//func TestEmailPattern(t *testing.T) {
//	testCase := []struct {
//		name  string
//		email string
//		match bool
//	}{
//		{
//			name:  "不带@",
//			email: "123456",
//			match: false,
//		},
//		{
//			name:  "带@但没后缀",
//			email: "123456@",
//			match: false,
//		},
//		{
//			name:  "合法邮箱",
//			email: "123456a@163.com",
//			match: true,
//		},
//	}
//
//	h := NewUserHandler(nil, nil)
//
//	for _, tc := range testCase {
//		t.Run(tc.name, func(t *testing.T) {
//			match, err := h.emailRegExp.MatchString(tc.email)
//			require.NoError(t, err)
//			assert.Equal(t, tc.match, match)
//		})
//	}
//}

//func TestMock(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//	userSvc := svcmocks.NewMockUserService(ctrl)
//	userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
//		Id:    1,
//		Email: "123@qq.com",
//	}).Return(errors.New("db 出错"))
//
//	err := userSvc.SignUp(context.Background(), domain.User{
//		Id:    1,
//		Email: "123@qq.com",
//	})
//	t.Log(err)
//}

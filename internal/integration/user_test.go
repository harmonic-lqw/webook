package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/internal/integration/startup"
	"webook/internal/web"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestUserHandler_SendSMSCode(t *testing.T) {
	const sendSMSCodeUrl = "/users/login_sms/code/send"
	rdb := startup.InitRedis()
	server := startup.InitWebServer()
	testCases := []struct {
		name string

		// 集成测试的关键点就在 before 和 after, 特别是在 after 中验证最终数据
		// before 准备数据
		before func(t *testing.T)
		// after 验证和删除数据
		after func(t *testing.T)

		phone string

		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {

			},
			// 希望 redis 中有验证码，而且有过期时间
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
				defer cancel()

				key := "phone_code:login:13912345678"
				code, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, len(code) > 0)
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > time.Minute*9+time.Second*40)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)

			},
			phone:    "13912345678",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},

		{
			name: "未输入手机号码",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			phone:    "",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "请输入手机号码",
			},
		},

		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
				defer cancel()

				key := "phone_code:login:13912345678"
				err := rdb.Set(ctx, key, "123456", time.Minute*10).Err()
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
				defer cancel()

				key := "phone_code:login:13912345678"
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)

			},
			phone:    "13912345678",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 4,
				Msg:  "短信发送太频繁，请稍后再试",
			},
		},

		{
			name: "系统错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
				defer cancel()

				key := "phone_code:login:13912345678"
				err := rdb.Set(ctx, key, "123456", 0).Err()
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
				defer cancel()

				key := "phone_code:login:13912345678"
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)

			},
			phone:    "13912345678",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			// 拿到 req 和 recorder
			req, err := http.NewRequest(http.MethodPost, sendSMSCodeUrl, bytes.NewReader([]byte(fmt.Sprintf(`{"phone": "%s"}`, tc.phone))))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// 发起调用
			server.ServeHTTP(recorder, req)

			// 断言
			assert.Equal(t, tc.wantCode, recorder.Code)
			// 测试 Bind， 因为此时返回 404 ，测到这里结束即可
			if tc.wantCode != http.StatusOK {
				return
			}
			var res web.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBody, res)

		})
	}
}

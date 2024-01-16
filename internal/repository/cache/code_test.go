package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/repository/cache/redismocks"
)

// mock Cmdable: mockgen -destination .\internal\repository\cache\redismocks\cmd_mock.go -package redismocks github.com/redis/go-redis/v9 Cmdable

func TestRedisCodeCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			// 可以发送验证码
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				//cmdV1 := redis.NewCmdResult(int64(0), nil)
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(0)) // 不是 -2 或 -1 即可
				cmd.SetErr(nil)

				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("test", "531111")}, []any{"123456"}).
					Return(cmd)

				return res
			},

			ctx:   context.Background(),
			biz:   "test",
			phone: "531111",
			code:  "123456",

			wantErr: nil,
		},

		{
			name: "调用 redis 出了问题",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(errors.New("redis Err"))

				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("test", "531111")}, []any{"123456"}).
					Return(cmd)

				return res
			},

			ctx:   context.Background(),
			biz:   "test",
			phone: "531111",
			code:  "123456",

			wantErr: errors.New("redis Err"),
		},

		{
			// res == -2
			name: "没有过期时间",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(-2))

				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("test", "531111")}, []any{"123456"}).
					Return(cmd)

				return res
			},

			ctx:   context.Background(),
			biz:   "test",
			phone: "531111",
			code:  "123456",

			wantErr: errors.New("验证码存在，但是没有过期时间"),
		},

		{
			// res == -1
			name: "验证码发送太快",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(-1))

				res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("test", "531111")}, []any{"123456"}).
					Return(cmd)

				return res
			},

			ctx:   context.Background(),
			biz:   "test",
			phone: "531111",
			code:  "123456",

			wantErr: ErrCodeSendTooMany,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cmd := tc.mock(ctrl)

			c := NewRedisCodeCache(cmd)

			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)

			assert.Equal(t, tc.wantErr, err)

		})
	}
}

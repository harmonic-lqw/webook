package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
)

func Test_userService_Login(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) repository.UserRepository

		// 预期输入
		ctx      context.Context
		email    string
		password string

		// 预期输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Email: "123@qq.com",
					// 仔细体会，这里拿到的密码，其实就是加密后的正确密码
					Password: "$2a$10$6IfqPo6qys.P.eurWwn5rupanHQ8pJVOy2Dih6WKoYrajn75Xgwaq",
					// 为了确保 wantUser 随便拿一个与登录业务无关的字段，这里是 phone
					Phone: "123456789",
				}, nil)

				return repo
			},

			// 用户输入的
			email:    "123@qq.com",
			password: "123456a@",

			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$6IfqPo6qys.P.eurWwn5rupanHQ8pJVOy2Dih6WKoYrajn75Xgwaq",
				Phone:    "123456789",
			},
			wantErr: nil,
		},

		{
			name: "用户未找到",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},

			// 用户输入的
			email:    "123@qq.com",
			password: "123456a@",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},

		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, errors.New("db Err"))
				return repo
			},

			// 用户输入的
			email:    "123@qq.com",
			password: "123456a@",

			wantUser: domain.User{},
			wantErr:  errors.New("db Err"),
		},

		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Email: "123@qq.com",
					// 仔细体会，这里拿到的密码，其实就是加密后的正确密码
					Password: "$2a$10$6IfqPo6qys.P.eurWwn5rupanHQ8pJVOy2Dih6WKoYrajn75Xgwaq",
					// 为了确保 wantUser 随便拿一个与登录业务无关的字段，这里是 phone
					Phone: "123456789",
				}, nil)

				return repo
			},

			// 用户输入的
			email:    "123@qq.com",
			password: "123456b@",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := tc.mock(ctrl)
			svc := NewUserService(userRepo)

			user, err := svc.Login(tc.ctx, tc.email, tc.password)

			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)

		})
	}
}

func TestPasswordEncrypt(t *testing.T) {
	password := []byte("123456a@")
	encrypted, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	println(string(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, []byte("123456a@"))
	assert.NoError(t, err)
}

func TestParseTime(t *testing.T) {
	birthday := "2017-12-05"
	dateFormat := "2006-01-02"

	birth, _ := time.ParseInLocation(dateFormat, birthday, time.Local)
	birthUnix := birth.UnixMilli()
	fmt.Println(birth)
	fmt.Println(birthUnix)

	birthTime := time.Unix(0, birthUnix*int64(time.Millisecond))
	// 将 time.Time 类型转换为字符串
	birthdaySting := birthTime.Format(dateFormat)

	fmt.Println(birthTime)
	fmt.Println(birthdaySting)

}

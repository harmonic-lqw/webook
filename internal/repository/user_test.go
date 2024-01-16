package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	cachemocks "webook/internal/repository/cache/mocks"
	"webook/internal/repository/dao"
	daomocks "webook/internal/repository/dao/mocks"
)

func TestCachedUserRepository_FindUserInfoById(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO)

		ctx context.Context
		uid int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，数据库查找成功",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindUserInfoById(gomock.Any(), uid).
					Return(
						dao.User{
							Id: uid,
							Email: sql.NullString{
								String: "123@qq.com",
								Valid:  true,
							},
							Password: "123456a@",
							Birthday: 100,
							AboutMe:  "关于我",
							Phone: sql.NullString{
								String: "15214661616",
								Valid:  true,
							},
							Ctime: 101,
							Utime: 102,
						}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456a@",
					Birthday: "1970-01-01", // Unix: 100
					AboutMe:  "关于我",
					Phone:    "15214661616",
					Ctime:    time.UnixMilli(101),
				}).Return(nil)
				return c, d

			},

			ctx: context.Background(),
			uid: int64(123),

			wantUser: domain.User{
				Id:       int64(123),
				Email:    "123@qq.com",
				Password: "123456a@",
				Birthday: "1970-01-01", // Unix: 100
				AboutMe:  "关于我",
				Phone:    "15214661616",
				Ctime:    time.UnixMilli(101),
			},
			wantErr: nil,
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{
					Id:       int64(123),
					Email:    "123@qq.com",
					Password: "123456a@",
					Birthday: "1970-01-01", // Unix: 100
					AboutMe:  "关于我",
					Phone:    "15214661616",
					Ctime:    time.UnixMilli(101),
				}, nil)
				return c, d
			},

			ctx: context.Background(),
			uid: int64(123),

			wantUser: domain.User{
				Id:       int64(123),
				Email:    "123@qq.com",
				Password: "123456a@",
				Birthday: "1970-01-01", // Unix: 100
				AboutMe:  "关于我",
				Phone:    "15214661616",
				Ctime:    time.UnixMilli(101),
			},
			wantErr: nil,
		},

		{
			name: "缓存未命中，数据库也未找到用户",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindUserInfoById(gomock.Any(), uid).
					Return(dao.User{}, dao.ErrRecordNotFound)
				return c, d

			},

			ctx: context.Background(),
			uid: int64(123),

			wantUser: domain.User{},
			wantErr:  dao.ErrRecordNotFound,
		},
		{
			name: "redis 写入错误",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindUserInfoById(gomock.Any(), uid).
					Return(
						dao.User{
							Id: uid,
							Email: sql.NullString{
								String: "123@qq.com",
								Valid:  true,
							},
							Password: "123456a@",
							Birthday: 100,
							AboutMe:  "关于我",
							Phone: sql.NullString{
								String: "15214661616",
								Valid:  true,
							},
							Ctime: 101,
							Utime: 102,
						}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456a@",
					Birthday: "1970-01-01", // Unix: 100
					AboutMe:  "关于我",
					Phone:    "15214661616",
					Ctime:    time.UnixMilli(101),
				}).Return(errors.New("redis Err"))
				return c, d

			},

			ctx: context.Background(),
			uid: int64(123),

			wantUser: domain.User{
				Id:       int64(123),
				Email:    "123@qq.com",
				Password: "123456a@",
				Birthday: "1970-01-01", // Unix: 100
				AboutMe:  "关于我",
				Phone:    "15214661616",
				Ctime:    time.UnixMilli(101),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userC, userD := tc.mock(ctrl)
			repo := NewCachedUserRepository(userD, userC)

			user, err := repo.FindUserInfoById(tc.ctx, tc.uid)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)

		})
	}

}

func TestToDomain(t *testing.T) {
	Birthday := int64(100)
	// 将 Unix 时间戳转换为 time.Time 类型
	birthTime := time.Unix(0, Birthday*int64(time.Millisecond))

	// 将 time.Time 类型转换为字符串
	birthdayString := birthTime.Format(time.DateOnly)

	t.Log(birthdayString)

}

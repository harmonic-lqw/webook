package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string
		mock func(t *testing.T) *sql.DB

		ctx  context.Context
		user User

		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				mockRes := sqlmock.NewResult(123, 1)
				// 这边要求传入的是 sql 的正则表达式
				mock.ExpectExec("INSERT INTO .*").
					WillReturnResult(mockRes)
				return db
			},

			ctx: context.Background(),
			user: User{
				NickName: "harmonic",
			},

			wantErr: nil,
		},

		{
			name: "邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				// 这边要求传入的是 sql 的正则表达式
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(&mysqlDriver.MySQLError{Number: 1062})
				return db
			},

			ctx: context.Background(),
			user: User{
				NickName: "harmonic",
			},

			wantErr: ErrDuplicateUser,
		},

		{
			name: "数据库错误",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				// 这边要求传入的是 sql 的正则表达式
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(errors.New("db Err"))
				return db
			},

			ctx: context.Background(),
			user: User{
				NickName: "harmonic",
			},

			wantErr: errors.New("db Err"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.mock(t)
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true, // 防止 GORM 在初始化时调用 show version
			}), &gorm.Config{
				DisableAutomaticPing:   true, // 防止 GORM 自动 ping 数据库
				SkipDefaultTransaction: true, // GORM 为了兼容不同数据库对 commit 的支持，会自动帮我们在 SQL 前后开启事务 (Begin...Commit)，即便是一个单一增删改语句， GORM 也会开启事务，这里要关闭
			})

			assert.NoError(t, err)

			dao := NewUserDAO(db)
			err = dao.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

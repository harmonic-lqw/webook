package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateUser  = errors.New("用户冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	Update(ctx context.Context, user User) error
	FindUserInfoById(ctx context.Context, userId int64) (User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

// User 注意区分和domain中的User的作用
type User struct {
	Id       int64          `gorm:"primaryKey, autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Password string

	Phone sql.NullString `gorm:"unique"`

	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64

	// 编辑字段
	NickName string `gorm:"type=varchar(128)"`
	Birthday int64
	AboutMe  string `gorm:"type=varchar(4096)"`
}

func (dao *GORMUserDAO) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	err := dao.db.WithContext(ctx).Create(&user).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		// 用户冲突
		if me.Number == duplicateErr {
			return ErrDuplicateUser
		}
	}
	return err
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("email=?", email).Find(&user).Error
	return user, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("phone=?", phone).Find(&user).Error
	return user, err
}

func (dao *GORMUserDAO) Update(ctx context.Context, user User) error {
	return dao.db.WithContext(ctx).Where("id=?", user.Id).Updates(&user).Error
}

func (dao *GORMUserDAO) FindUserInfoById(ctx context.Context, userId int64) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("id=?", userId).Find(&user).Error
	return user, err
}

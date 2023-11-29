package repository

import (
	"context"
	"github.com/gin-gonic/gin"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

const (
	dateFormat = "2006-01-02"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error {
	var user = dao.User{
		Email:    u.Email,
		Password: u.Password,
	}
	return repo.dao.Insert(ctx, user)

}

func (repo *UserRepository) FindByEmail(ctx *gin.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *UserRepository) toDomain(u dao.User) domain.User {
	var birthdaySting string
	if u.Birthday != int64(0) {
		// 将 Unix 时间戳转换为 time.Time 类型
		birthTime := time.Unix(0, u.Birthday*int64(time.Millisecond))

		// 将 time.Time 类型转换为字符串
		birthdaySting = birthTime.Format(dateFormat)
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		NickName: u.NickName,
		AboutMe:  u.AboutMe,
		Birthday: birthdaySting,
	}
}

func (repo *UserRepository) EditUserInfo(ctx *gin.Context, userID int64, name string, birthday string, me string) error {
	birth, _ := time.ParseInLocation(dateFormat, birthday, time.Local)
	birthUnix := birth.UnixMilli()

	var user = dao.User{
		Id:       userID,
		NickName: name,
		Birthday: birthUnix,
		AboutMe:  me,
	}
	return repo.dao.Update(ctx, user)
}

func (repo *UserRepository) FindUserInfoById(ctx *gin.Context, userID int64) (domain.User, error) {
	u, err := repo.dao.FindUserInfoById(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

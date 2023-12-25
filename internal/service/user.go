package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"webook/internal/domain"
	"webook/internal/repository"
)

var (
	ErrDuplicateUser         = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户不存在或密码错误")
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx *gin.Context, email string, password string) (domain.User, error)
	EditUserInfo(ctx *gin.Context, userID int64, name string, birthday string, me string) error
	GetUserInfo(ctx *gin.Context, userID int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// SignUp 服务层面叫SignUp
func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	// 加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)

	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx *gin.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *userService) EditUserInfo(ctx *gin.Context, userID int64, name string, birthday string, me string) error {
	return svc.repo.EditUserInfo(ctx, userID, name, birthday, me)
}

func (svc *userService) GetUserInfo(ctx *gin.Context, userID int64) (domain.User, error) {
	u, err := svc.repo.FindUserInfoById(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 先去查找一下，因为我们认为大部分用户是已存在的用户
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != nil && err != repository.ErrUserNotFound {
		// 此时有两种情况
		// 1. err == nil，u 找到了
		// 2. err != nil, 系统错误
		return u, err
	}
	// 用户没找到 err == repository.ErrUserNotFound
	// 触发注册
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})

	// 此时两种可能
	// 一种是 err != nil, 系统错误
	// 一种是 err 恰好是唯一索引冲突（这里就是phone）

	// 第一种：系统错误
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	// 第二种：也代表用户存在，因此此时和 err == nil 一样，直接返回即可
	// （不过会有主从延迟的问题）
	return svc.repo.FindByPhone(ctx, phone)
}

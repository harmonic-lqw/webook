package repository

import (
	"context"
	"database/sql"
	"log"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateUser
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	EditUserInfo(ctx context.Context, userID int64, name string, birthday string, me string) error
	FindUserInfoById(ctx context.Context, userID int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCachedUserRepository(d dao.UserDAO, c cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   d,
		cache: c,
	}
}

func (repo *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.toEntity(u))

}

func (repo *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *CachedUserRepository) EditUserInfo(ctx context.Context, userID int64, name string, birthday string, me string) error {
	birthUnix := repo.timeStoUnix(birthday)
	var user = dao.User{
		Id:       userID,
		NickName: name,
		Birthday: birthUnix,
		AboutMe:  me,
	}
	err := repo.dao.Update(ctx, user)
	if err != nil {
		return err
	}
	du := repo.toDomain(user)
	// 刷新缓存
	go func() {
		err = repo.cache.Set(ctx, du)
		if err != nil {
			// 记录日志
		}
	}()
	return nil
}

func (repo *CachedUserRepository) FindUserInfoById(ctx context.Context, userID int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, userID) // 性能优化：redis 缓存方案

	// err 为 nil ，说明 redis 中有数据且成功查询到，直接返回即可
	if err == nil {
		return du, nil
	}

	// err 不为nil，就要查询数据库
	// err 有两种可能
	// 1. key 不存在，说明 redis 正常工作
	// 2. redis 有问题。有可能是网络问题，也可能是 redis 本身就崩溃了

	u, err := repo.dao.FindUserInfoById(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	du = repo.toDomain(u)

	// 刷新 redis 的异步写法，能够提高一些查询性能
	//go func() {
	//	err = repo.cache.Set(ctx, du) // 性能优化：redis 缓存方案
	//	if err != nil {
	//		log.Println(err)
	//	}
	//}()

	// 刷新 redis
	err = repo.cache.Set(ctx, du) // 性能优化：redis 缓存方案
	if err != nil {
		// 网络崩了，也可能是 redis 崩了
		log.Println(err)
	}

	return du, nil
}

func (repo *CachedUserRepository) FindUserInfoByIdV1(ctx context.Context, userID int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, userID) // 性能优化：redis 缓存方案

	switch err {
	case nil:
		return du, nil
	case cache.ErrKeyNotExist:
		u, err := repo.dao.FindUserInfoById(ctx, userID)
		if err != nil {
			return domain.User{}, err
		}
		du = repo.toDomain(u)

		// 刷新 redis
		err = repo.cache.Set(ctx, du) // 性能优化：redis 缓存方案
		if err != nil {
			// 网络崩了，也可能是 redis 崩了
			log.Println(err)
		}
		return du, nil
	default: // redis 运作不正常，不需要查询数据库，防止数据库压力过大
		// 降级写法
		return domain.User{}, err

	}

}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *CachedUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	u, err := repo.dao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *CachedUserRepository) timeStoUnix(timeS string) int64 {
	// 将字符串转为 time.Time 类型
	birth, _ := time.ParseInLocation(time.DateOnly, timeS, time.Local)
	// 将 time.Time 转为 Unix 时间戳
	birthUnix := birth.UnixMilli()
	return birthUnix
}

func (repo *CachedUserRepository) toEntity(u domain.User) dao.User {
	birthUnix := repo.timeStoUnix(u.Birthday)
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
		Birthday: birthUnix,
		AboutMe:  u.AboutMe,
		NickName: u.NickName,
		WechatOpenId: sql.NullString{
			String: u.WechatInfo.OpenId,
			Valid:  u.WechatInfo.OpenId != "",
		},
		WechatUnionId: sql.NullString{
			String: u.WechatInfo.UnionId,
			Valid:  u.WechatInfo.UnionId != "",
		},
	}
}

func (repo *CachedUserRepository) toDomain(u dao.User) domain.User {
	var birthdayString string
	if u.Birthday != int64(0) {
		// 将 Unix 时间戳转换为 time.Time 类型
		birthTime := time.Unix(0, u.Birthday*int64(time.Millisecond))

		// 将 time.Time 类型转换为字符串
		birthdayString = birthTime.Format(time.DateOnly)
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		NickName: u.NickName,
		AboutMe:  u.AboutMe,
		Birthday: birthdayString,
		Ctime:    time.UnixMilli(u.Ctime),
		WechatInfo: domain.WechatInfo{
			OpenId:  u.WechatOpenId.String,
			UnionId: u.WechatUnionId.String,
		},
	}
}

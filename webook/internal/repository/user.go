package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrUserDuplicate      = dao.ErrUserDuplicate
	ErrUserNotFound       = dao.ErrUserNotFound
	ErrInvalidData        = dao.ErrInvalidData
	ErrRecordNotFound     = dao.ErrRecordNotFound
	ErrCodeSendTooMany    = dao.ErrCodeSendTooMany
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
)

type UserRepository interface {
	Create(c context.Context, u domain.User) error
	// Update 更新数据，只有非 0 值才会更新
	Update(c context.Context, u domain.User) error
	FindById(c context.Context, uid int64) (domain.User, error)
	FindByEmail(c context.Context, email string) (domain.User, error)
	FindByPhone(c context.Context, phone string) (domain.User, error)
	FindByWechat(c context.Context, openID string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, c cache.UserCache) UserRepository { // 所有东西都从外面传进来，不初始化
	return &CachedUserRepository{
		dao:   dao,
		cache: c,
	}
}

func (r *CachedUserRepository) Create(c context.Context, u domain.User) error {
	return r.dao.Insert(c, r.domainToEntity(u))
}

func (r *CachedUserRepository) Update(c context.Context, u domain.User) error {
	if err := r.dao.UpdateNonZeroFields(c, r.domainToEntity(u)); err != nil {
		return fmt.Errorf("error" + err.Error())
	}
	return r.cache.Delete(c, u.Id)
}

func (r *CachedUserRepository) FindById(c context.Context, uid int64) (domain.User, error) {
	// 先从 cache 里面找，再从 dao 里面找， 找到了回写 cache
	// SELECT * FROM `users` WHERE `id`=?
	u, err := r.cache.Get(c, uid)
	if err == nil {
		// 必然是有数据
		return u, nil
	}
	uu, err := r.dao.FindByUserId(c, uid)
	if err != nil {
		return domain.User{}, fmt.Errorf("id can not be found")
	}
	u = r.entityToDomain(uu)
	// _ = r.cache.Set(c, u)
	// 一致性
	go func() {
		err = r.cache.Set(c, u)
		if err != nil {
			// 打日志，做监控，防redis崩
			log.Fatalf("error. %v\n", err)
		}
	}()
	// 选加载数据库 ---- 但要做好兜底，redis一旦崩了，要保护住数据库------解决方案：数据库做限流。/给redis做两个集群，互相用。
	// 选不加载 ---- 用户体验稍差
	return u, nil
}

func (r *CachedUserRepository) FindByEmail(c context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email`=?
	u, err := r.dao.FindByEmail(c, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("email can not be found. %w" + err.Error())
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) FindByPhone(c context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByUserPhone(c, phone)
	if err != nil {
		return domain.User{}, fmt.Errorf("phone not found. %w\n" + err.Error())
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) FindByWechat(c context.Context, openID string) (domain.User, error) {
	u, err := r.dao.FindByWechat(c, openID)
	if err != nil {
		return domain.User{}, fmt.Errorf("openID not found. %w\n" + err.Error())
	}
	return r.entityToDomain(u), nil
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
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
		Ctime:    u.Ctime.UnixMilli(),
		WechatOpenID: sql.NullString{
			String: u.WechatInfo.OpenID,
			Valid:  u.WechatInfo.OpenID != "",
		},
		WechatUnionID: sql.NullString{
			String: u.WechatInfo.UnionID,
			Valid:  u.WechatInfo.UnionID != "",
		},
	}
}

func (r *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
		WechatInfo: domain.WechatInfo{
			OpenID:  u.WechatOpenID.String,
			UnionID: u.WechatUnionID.String,
		},
	}
}

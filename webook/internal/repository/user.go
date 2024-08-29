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
	ErrUserDuplicate  = dao.ErrUserDuplicate
	ErrUserNotFound   = dao.ErrUserNotFound
	ErrInvalidData    = dao.ErrInvalidData
	ErrRecordNotFound = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, c *cache.UserCache) *UserRepository { // 所有东西都从外面传进来，不初始化
	return &UserRepository{
		dao:   dao,
		cache: c,
	}
}

func (r *UserRepository) Create(c context.Context, u domain.User) error {
	return r.dao.Insert(c, r.domainToEntity(u))
}

func (r *UserRepository) Update(c context.Context, uid int64, u domain.User) error {
	return r.dao.Update(c, uid, dao.User{
		NickName: u.NickName,
		Birthday: u.Birthday,
		Bio:      u.Bio,
	})
}

func (r *UserRepository) FindById(c context.Context, uid int64) (domain.User, error) {
	// 先从 cache 里面找，再从 dao 里面找， 找到了回写 cache
	// SELECT * FROM `users` WHERE `id`=?
	u, err := r.cache.Get(c, uid)
	if err == nil {
		// 必然是有数据
		return u, nil
	}
	uu, err := r.dao.FindByUserId(c, uid)
	if err != nil {
		return domain.User{}, fmt.Errorf("id can not be found. %w", err)
	}
	u = r.entityToDomain(uu)
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

func (r *UserRepository) FindByEmail(c context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email`=?
	u, err := r.dao.FindByEmail(c, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("email can not be found. %w", err)
	}
	return r.entityToDomain(u), nil
}

func (r *UserRepository) FindByPhone(c context.Context, phone string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email`=?
	u, err := r.dao.FindByUserPhone(c, phone)
	if err != nil {
		return domain.User{}, fmt.Errorf("phone not found. %w", err)
	}
	return r.entityToDomain(u), nil
}

func (r *UserRepository) domainToEntity(u domain.User) dao.User {
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
	}
}

func (r *UserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
	}
}

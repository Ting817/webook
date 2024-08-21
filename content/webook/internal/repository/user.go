package repository

import (
	"context"
	"fmt"

	"junior-engineer-training/content/webook/internal/domain"
	"junior-engineer-training/content/webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
	ErrInvalidData        = dao.ErrInvalidData
	ErrRecordNotFound     = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository { // 所有东西都从外面传进来，不初始化
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) Create(c context.Context, u domain.User) error {
	return r.dao.Insert(c, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) Update(c context.Context, uid interface{}, u domain.User) error {
	return r.dao.Update(c, uid, dao.User{
		NickName: u.NickName,
		Birthday: u.Birthday,
		Bio:      u.Bio,
	})
}

func (r *UserRepository) FindById(c context.Context, uid interface{}) (domain.User, error) {
	// 先从 cache 里面找，再从 dao 里面找， 找到了回写 cache
	// SELECT * FROM `users` WHERE `id`=?
	u, err := r.dao.FindByUserId(c, uid)
	if err != nil {
		return domain.User{}, fmt.Errorf("id can not be found. %w", err)
	}
	return domain.User{
		Email:    u.Email,
		Password: u.Password,
		Id:       u.Id,
		NickName: u.NickName,
		Birthday: u.Birthday,
		Bio:      u.Bio,
	}, nil
}

func (r *UserRepository) FindByEmail(c context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email`=?
	u, err := r.dao.FindByEmail(c, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("email can not be found. %w", err)
	}
	return domain.User{
		Email:    u.Email,
		Password: u.Password,
		Id:       u.Id,
		NickName: u.NickName,
		Birthday: u.Birthday,
		Bio:      u.Bio,
	}, nil
}

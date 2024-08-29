package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"webook/internal/domain"
	"webook/internal/repository"
)

var (
	ErrUserDuplicate         = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("user or password error")
	ErrInvalidData           = repository.ErrInvalidData
	ErrRecordNotFound        = repository.ErrRecordNotFound
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(c context.Context, u domain.User) error {
	// 加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(c, u)
}

func (svc *UserService) Login(c context.Context, email, password string) (domain.User, error) {
	// 先找用户
	u, err := svc.repo.FindByEmail(c, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	if err != nil {
		return domain.User{}, err
	}

	// 比较密码了
	if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	return u, nil
}

func (svc *UserService) Edit(c context.Context, uid int64, u domain.User) error {
	return svc.repo.Update(c, uid, u)
}

func (svc *UserService) Profile(c context.Context, uid int64) (domain.User, error) {
	u, err := svc.repo.FindById(c, uid)
	if err != nil {
		return domain.User{}, fmt.Errorf("nothing found")
	}
	return u, nil
}

func (svc *UserService) FindOrCreate(c context.Context, phone string) (domain.User, error) {
	// 快路径
	u, err := svc.repo.FindByPhone(c, phone)
	if err != nil {
		return u, fmt.Errorf("user find by phone failed. %w\n", err)
	}
	if c.Value("降级") == "true" {
		return domain.User{}, fmt.Errorf("系统降级了. %w\n", err)
	}
	// 慢路径
	// 如果没有这个用户
	u = domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(c, u)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return u, fmt.Errorf("create user by phone failed. %w\n", err)
	}

	// 这里会遇到主从延迟的问题
	return svc.repo.FindByPhone(c, phone)
}

package service

import (
	"context"
	"errors"
	"fmt"
	"webook/pkg/logger"

	"golang.org/x/crypto/bcrypt"

	"webook/internal/domain"
	"webook/internal/repository"
)

var (
	ErrUserDuplicate         = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("user or password error")
	ErrInvalidData           = repository.ErrInvalidData
	ErrRecordNotFound        = repository.ErrRecordNotFound
	ErrCodeSendTooMany       = repository.ErrCodeSendTooMany
)

type UserService interface {
	SignUp(c context.Context, u domain.User) error
	FindOrCreate(c context.Context, phone string) (domain.User, error)
	Login(c context.Context, email, password string) (domain.User, error)
	Edit(c context.Context, uid int64, u domain.User) error
	Profile(c context.Context, uid int64) (domain.User, error)
	// UpdateNonSensitiveInfo 更新非敏感数据
	UpdateNonSensitiveInfo(c context.Context, user domain.User) error
	FindOrCreateByWechat(c context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
	l    logger.LoggerV1
}

func NewUserService(repo repository.UserRepository, l logger.LoggerV1) UserService {
	return &userService{
		repo: repo,
		l:    l,
	}
}

func (svc *userService) SignUp(c context.Context, u domain.User) error {
	// 加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(c, u)
}

func (svc *userService) FindOrCreate(c context.Context, phone string) (domain.User, error) {
	// 快路径
	u, err := svc.repo.FindByPhone(c, phone)
	if err != nil {
		return u, fmt.Errorf("user find by phone failed. %w\n", err)
	}
	//zap.L().Info("user not sign up", zap.String("phone", phone)) // 手机号要先脱敏
	svc.l.Info("user not sign up", logger.Field{
		Key:   "phone",
		Value: phone,
	})
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

func (svc *userService) FindOrCreateByWechat(c context.Context, info domain.WechatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechat(c, info.OpenID)
	if err != nil {
		return u, fmt.Errorf("openID find by wechat failed. %w\n", err)
	}

	u = domain.User{
		WechatInfo: info,
	}
	err = svc.repo.Create(c, u)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return u, fmt.Errorf("create user by phone failed. %w\n", err)
	}

	// 这里会遇到主从延迟的问题
	return svc.repo.FindByWechat(c, info.OpenID)
}

func (svc *userService) Login(c context.Context, email, password string) (domain.User, error) {
	// 先找用户
	u, err := svc.repo.FindByEmail(c, email)
	if err == repository.ErrUserNotFound {
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

func (svc *userService) Edit(c context.Context, uid int64, u domain.User) error {
	return svc.repo.Update(c, u)
}

func (svc *userService) Profile(c context.Context, uid int64) (domain.User, error) {
	u, err := svc.repo.FindById(c, uid)
	if err != nil {
		return domain.User{}, fmt.Errorf("nothing found")
	}
	return u, nil
}

func (svc *userService) UpdateNonSensitiveInfo(c context.Context, u domain.User) error {
	// 写法1
	// 这种是简单的写法，依赖与 Web 层保证没有敏感数据被修改
	// 也就是说，你的基本假设是前端传过来的数据就是不会修改 Email，Phone 之类的信息的。
	// return svc.repo.Update(ctx, user)

	// 写法2
	// 这种是复杂写法，依赖于 repository 中更新会忽略 0 值
	// 这个转换的意义在于，你在 service 层面上维护住了什么是敏感字段这个语义
	u.Email = ""
	u.Phone = ""
	u.Password = ""
	return svc.repo.Update(c, u)
}

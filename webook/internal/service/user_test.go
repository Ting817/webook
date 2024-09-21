package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"webook/pkg/logger"

	"github.com/go-playground/assert/v2"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
)

func Test_userService_Login(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		c        context.Context
		email    string
		password string
		l        logger.LoggerV1
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "login success!",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Email:    "123@qq.com",
					Password: "$2a$10$UXxR5.csk4.R8B5.Xy2tLuShAyMllZy9BJ61LP2JwUP44k5k3hYAC",
					Phone:    "1233455677",
					Ctime:    now,
				}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "hello@world123",

			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$UXxR5.csk4.R8B5.Xy2tLuShAyMllZy9BJ61LP2JwUP44k5k3hYAC",
				Phone:    "1233455677",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123@qq.com",
			password: "hello@world123",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "mock db error",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, errors.New("mock db error"))
				return repo
			},
			email:    "123@qq.com",
			password: "hello@world123",
			wantUser: domain.User{},
			wantErr:  errors.New("mock db error"),
		},
		{
			name: "Invalid Password",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, ErrInvalidUserOrPassword)
				return repo
			},
			email:    "123@qq.com",
			password: "hello@world123",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewUserService(tt.mock(ctrl), tt.l)
			u, err := svc.Login(tt.c, tt.email, tt.password)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantUser, u)
		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("hello@world123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}

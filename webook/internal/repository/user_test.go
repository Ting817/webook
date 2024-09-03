package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"go.uber.org/mock/gomock"

	"webook/internal/domain"
	"webook/internal/repository/cache"
	cachemocks "webook/internal/repository/cache/mocks"
	"webook/internal/repository/dao"
	daomocks "webook/internal/repository/dao/mocks"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	now := time.Now()
	now = time.UnixMilli(now.UnixMilli()) // 去掉毫秒以外的部分
	tests := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		c        context.Context
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "cache not hit, query succeeded",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(domain.User{}, errors.New("id not found"))

				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindByUserId(gomock.Any(), int64(123)).Return(dao.User{
					Id: 123,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Phone: sql.NullString{
						String: "12345678",
						Valid:  true,
					},
					Password: "xxx",
					Ctime:    now.UnixMilli(),
					Utime:    now.UnixMilli(),
				}, nil)

				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "xxx",
					Phone:    "12345678",
					Ctime:    now,
				}).Return(nil)

				return d, c
			},
			c:  context.Background(),
			id: 123,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "xxx",
				Phone:    "12345678",
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "user found, direct hit cache",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				// 注意，我们传入的是 int64，
				// 所以要做一个显式的转换，不然默认 12 是 int 类型
				c.EXPECT().Get(gomock.Any(), int64(12)).
					Return(domain.User{
						Id:       12,
						Email:    "123@qq.com",
						Password: "123456",
						Phone:    "15212345678",
						Ctime:    now,
					}, nil)
				return d, c
			},

			c:  context.Background(),
			id: 12,

			wantUser: domain.User{
				Id:       12,
				Email:    "123@qq.com",
				Password: "123456",
				Phone:    "15212345678",
				Ctime:    now,
			},
		},
		{
			name: "user not found",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				// 注意这边，我们传入的是 int64，
				// 所以要做一个显式的转换，不然默认 12 是 int 类型
				c.EXPECT().Get(gomock.Any(), int64(12)).Return(domain.User{}, errors.New("id can not be found"))
				d.EXPECT().FindByUserId(gomock.Any(), int64(12)).Return(dao.User{}, dao.ErrUserNotFound)
				return d, c
			},

			c:       context.Background(),
			id:      12,
			wantErr: errors.New("id can not be found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ud, uc := tt.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			u, err := repo.FindById(tt.c, tt.id)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantUser, u)
			// 为了实现 go func()
			time.Sleep(time.Second)
		})
	}
}

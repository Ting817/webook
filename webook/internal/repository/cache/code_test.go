package cache

import (
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/mock/gomock"

	"webook/internal/repository/cache/redismocks"
)

func TestRedisCodeCache_Set(t *testing.T) {
	tests := []struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable

		// 输入
		c     context.Context
		biz   string
		phone string
		code  string

		// 输出
		wantErr error
	}{
		{
			name: "code set success",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				// res.SetErr(nil)
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code: login:123"}, []any{"12345"}).Return(res)
				return cmd
			},
			c:       context.Background(),
			biz:     "login",
			phone:   "123",
			code:    "12345",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			nc := NewCodeCache(tt.mock(ctrl))
			err := nc.Set(tt.c, tt.biz, tt.phone, tt.code)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

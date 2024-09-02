package web

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"webook/internal/domain"
	"webook/internal/service"
	svcmocks "webook/internal/service/mocks"
)

func TestUserHandler_SignUp(t *testing.T) {
	tests := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "sign up success!",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(nil) // 注册成功是 return nil
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello@world123",
	"confirmPassword": "hello@world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "sign up success!",
		},
		{
			name: "The email format is incorrect.",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
{
	"email": "123",
	"password": "hello@world123",
	"confirmPassword": "hello@world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "The email format is incorrect.",
		},
		{
			name: "The password must be longer than 8 characters and include both numbers and special symbols.",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello",
	"confirmPassword": "hello"
}
`,
			wantCode: http.StatusOK,
			wantBody: "The password must be longer than 8 characters and include both numbers and special symbols.",
		},
		{
			name: "passwords do not match.",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello@world1234",
	"confirmPassword": "hello@wod123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "passwords do not match.",
		},
		{
			name: "email conflict.",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(service.ErrUserDuplicate)
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello@world123",
	"confirmPassword": "hello@world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "email conflict.",
		},
		{
			name: "system error.",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(errors.New("system error"))
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello@world123",
	"confirmPassword": "hello@world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "system error.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			usersvc := tt.mock(ctrl)
			h := NewUserHandler(usersvc, nil)
			server := gin.Default()
			h.RegisterRoutes(server)
			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tt.reqBody)))
			require.NoError(t, err)
			// 数据是json格式
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			t.Log(resp)

			server.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantCode, resp.Code)
			assert.Equal(t, tt.wantBody, resp.Body.String())
		})
	}
}

func TestMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usersvc := svcmocks.NewMockUserService(ctrl)

	usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).
		Return(errors.New("mock error"))

	err := usersvc.SignUp(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)
}

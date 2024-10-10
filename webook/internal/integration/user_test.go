package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/internal/integration/startup"
	"webook/internal/web"
)

func TestUserHandler_e2e_SendLoginSMSCode(t *testing.T) {
	server := startup.InitWebServer()
	rdb := startup.InitRedis()
	tests := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		reqBody string

		wantCode int
		wantBody web.Result
	}{
		{
			name: "send code success!",
			before: func(t *testing.T) {
				// 不需要 即 Redis 什么数据也没有
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理数据
				val, err := rdb.GetDel(ctx, "phone_code:login:123456789").Result()
				cancel()
				assert.NoError(t, err)
				// 验证码是6位
				assert.True(t, len(val) == 6)
			},
			reqBody: `
{
	"phone": "123456789"
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Msg: "send code success!",
			},
		},
		{
			name: "code send too many",
			before: func(t *testing.T) {
				// 已有一个验证码
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				_, err := rdb.Set(ctx, "phone_code:login:123456789", "123456", time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理数据
				val, err := rdb.Get(ctx, "phone_code:login:123456789").Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
			},
			reqBody: `
{
	"phone": "123456789"
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 4,
				Msg:  "code send too many, please try it again later",
			},
		},
		{
			name: "system error",
			before: func(t *testing.T) {
				// 已有一个验证码, 但没有过期时间
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				_, err := rdb.Set(ctx, "phone_code:login:123456789", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				// 清理数据
				val, err := rdb.Set(ctx, "phone_code:login:123456789", "123456", time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)
				assert.Equal(t, "OK", val)
			},
			reqBody: `
{
	"phone": "123456789"
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 5,
				Msg:  "system error",
			},
		},
		{
			name: "phone is empty",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody: `
{
	"phone": ""
}
`,
			wantCode: 200,
			wantBody: web.Result{
				Code: 4,
				Msg:  "phone input error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(t)
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewBuffer([]byte(tt.reqBody)))
			assert.NoError(t, err)
			// 数据是json格式
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			t.Log(resp)

			server.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantCode, resp.Code)

			var webRes web.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantBody, webRes)
			tt.after(t)
		})
	}
}

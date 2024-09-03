package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
	key string
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}

// Send 其中 biz 必须是线下申请的一个代表业务方的 token
func (s *SMSService) Send(c context.Context, biz string, args []string, numbers ...string) error {
	var tc Claims
	// 如果解析成功，说明是对应的业务方
	// 没有 error 说明，token 是我发的
	token, err := jwt.ParseWithClaims(biz, &tc, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return fmt.Errorf("get token error %w", err)
	}
	if !token.Valid {
		return fmt.Errorf("token is invalid")
	}
	return s.svc.Send(c, tc.Tpl, args, numbers...)
}

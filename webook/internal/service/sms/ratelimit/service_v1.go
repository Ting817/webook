package ratelimit

import (
	"context"
	"fmt"
	"webook/internal/service/sms"
	"webook/pkg/ratelimit"
)

type ServiceV1 struct {
	sms.Service // 组合法 用户可直接访问Service, 绕开装饰器本身，可选择性实现/装饰需要的
	limiter     ratelimit.Limiter
}

func NewServiceV1(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &Service{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *ServiceV1) Send(c context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(c, "sms:tencent")
	if err != nil {
		// 若系统错误，到底要不要限流？
		// 情况1：可限流：保守策略 你的下游很坑
		// 情况2： 可不限：下游很强，业务可用性要求很高，尽量容错策略
		return fmt.Errorf("problems with SMS services determining whether to limit traffic，%w", err)
	}
	if limited {
		return errLimited // 不到万不得已 不要做成公开的错误
	}
	// 加新特性
	err = s.Service.Send(c, tpl, args, numbers...)
	return err
}

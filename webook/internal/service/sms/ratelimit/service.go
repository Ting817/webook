package ratelimit

import (
	"context"
	"fmt"
	"webook/internal/service/sms"
	"webook/pkg/ratelimit"
)

var errLimited = fmt.Errorf("triggered a current limit")

type Service struct {
	svc     sms.Service // 被装饰的，必须实现Service所有方法，可以有效阻止用户绕开装饰器
	limiter ratelimit.Limiter
}

func NewService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &Service{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *Service) Send(c context.Context, tpl string, args []string, numbers ...string) error {
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
	err = s.svc.Send(c, tpl, args, numbers...)
	return err
}

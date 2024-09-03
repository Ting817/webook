package retryable

import (
	"context"
	"fmt"
	"webook/internal/service/sms"
)

type Service struct {
	svc      sms.Service
	retryMax int
}

// Send 重试机制
func (s Service) Send(c context.Context, biz string, args []string, numbers ...string) error {
	err := s.svc.Send(c, biz, args, numbers...)
	cnt := 1
	for err != nil && cnt < s.retryMax {
		err = s.svc.Send(c, biz, args, numbers...)
		if err == nil {
			return nil
		}
		cnt++
	}
	return fmt.Errorf("retry failed")
}

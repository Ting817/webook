package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"webook/internal/service/sms"
)

// failover (轮询)策略，如果sms出现错误了，就直接换一个服务商，进行重试

type FailoverService struct {
	svcs []sms.Service
	idx  uint64
}

func NewFailoverService(svcs []sms.Service) sms.Service {
	return &FailoverService{
		svcs: svcs,
	}
}

func (f *FailoverService) Send(c context.Context, tpl string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(c, tpl, args, numbers...)
		// 发送成功
		if err == nil {
			return nil
		}
		// 正常是输出日志 做好监控
		log.Println(err)
	}
	return errors.New("all svcs are failed")
}

func (f *FailoverService) SendV1(c context.Context, tpl string, args []string, numbers ...string) error {
	// 从下标的下一位开始，即从下一个服务商开始
	// 原子读，你不会读到修改了一半的数据
	// 原子操作是轻量级并发工具，一种并发优化的思路，注意原子操作操作的都是指针
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[int(i%length)]
		err := svc.Send(c, tpl, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			// 调用者设置的超时时间到了，主动取消了进程
			return err
		default:
			// 输出日志

		}
	}
	return errors.New("all svcs are failed")
}

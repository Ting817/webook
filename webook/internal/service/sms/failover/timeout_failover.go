package failover

import (
	"context"
	"sync/atomic"
	"webook/internal/service/sms"
)

type TimeoutFailoverService struct {
	svcs []sms.Service
	idx  int32
	// 连接超时的个数
	cnt int32
	// 阈值 连续超时超过这个数，就要切换
	threshold int32
}

func NewTimeoutFailoverService() sms.Service {
	return &TimeoutFailoverService{}
}

// Send 非严谨的“连续 N 个超时就切换”
func (t *TimeoutFailoverService) Send(c context.Context, tpl string, args []string, numbers ...string) error {
	idx := atomic.AddInt32(&t.idx, 1)
	cnt := atomic.LoadInt32(&t.cnt)
	if cnt > t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs)) // 取余，防止溢出
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 成功往后挪了一位
			atomic.StoreInt32(&t.cnt, 0)
		}
		// else 出现并发，别人换成功了

		// 两种写法 都可以
		idx = newIdx
		//idx = atomic.LoadInt32(&t.idx)
	}

	svc := t.svcs[idx]
	err := svc.Send(c, tpl, args, numbers...)
	switch err {
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
		return err
	case nil:
		// 连续状态被打断
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	default:
		// 可考虑：1.超时错误，可能时偶发的，尽量再试试；2.非超时，直接下一个；3.等等...、
		return err
	}
}

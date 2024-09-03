package ratelimit

import (
	"context"
)

// Limiter limit: 有没有触发限流。 key: 限流对象
// bool 为 true 即表示要限流
// error: 限流器本身有没有错误
type Limiter interface {
	Limit(c context.Context, key string) (bool, error)
}

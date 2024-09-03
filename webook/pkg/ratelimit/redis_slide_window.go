package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed slide_window.lua
var luaSlideWindow string

// RedisSlidingWindowLimiter Redis 上的滑动窗口算法限流器的实现
type RedisSlidingWindowLimiter struct {
	cmd      redis.Cmdable
	interval time.Duration // 窗口大小
	rate     int           // 阈值
}

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &RedisSlidingWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisSlidingWindowLimiter) Limit(c context.Context, key string) (bool, error) {
	return r.cmd.Eval(c, luaSlideWindow, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}

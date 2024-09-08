package cache

import (
	"context"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"time"
)

type LocalCodeCache struct {
	cache      *lru.Cache
	lock       sync.Mutex
	expiration time.Duration
}

func NewLocalCodeCache(c *lru.Cache, expiration time.Duration) *LocalCodeCache {
	return &LocalCodeCache{
		cache:      c,
		expiration: expiration,
	}
}

func (l *LocalCodeCache) Set(ctx context.Context, biz string, phone string, code string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	// 这里可以考虑用读写锁来优化，但是效果不会很好
	// 因为你可以预期，大部分时候是要走到写锁里面的

	// 我选用的本地缓存，很不幸的是，没有获得过期时间的接口，所以都是自己维持了一个过期时间字段
	key := l.key(biz, phone)
	now := time.Now()
	val, ok := l.cache.Get(key)
	if !ok {
		// 说明没有验证码
		l.cache.Add(key, codeItem{
			code:   code,
			cnt:    3,
			expire: now.Add(l.expiration),
		})
		return nil
	}
	itm, ok := val.(codeItem)
	if !ok {
		// 理论上来说这是不可能的
		return errors.New("系统错误")
	}
	if itm.expire.Sub(now) > time.Minute*9 {
		// 不到一分钟
		return errors.New("Code send too many. ")
	}
	// 重发
	l.cache.Add(key, codeItem{
		code:   code,
		cnt:    3,
		expire: now.Add(l.expiration),
	})
	return nil
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	key := l.key(biz, phone)
	val, ok := l.cache.Get(key)
	if !ok {
		// 都没发验证码
		return false, errors.New("key not exists")
	}
	itm, ok := val.(codeItem)
	if !ok {
		// 理论上来说这是不可能的
		return false, errors.New("system error")
	}
	if itm.cnt <= 0 {
		return false, errors.New("code send too many")
	}
	itm.cnt--
	return itm.code == inputCode, nil
}

func (l *LocalCodeCache) key(biz string, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type codeItem struct {
	code string
	// 可验证次数
	cnt int
	// 过期时间
	expire time.Time
}

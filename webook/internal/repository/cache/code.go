package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
)

//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache interface {
	Set(c context.Context, biz, phone, code string) error
	Verify(c context.Context, biz, phone, inputCode string) (bool, error)
}

type RedisCodeCache struct {
	client redis.Cmdable
}

// 其实Go的最佳实践是返回具体类型，而不是返回接口。此处用了wire，所以用接口

func NewRedisCodeCache(client redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

func (cc *RedisCodeCache) Set(c context.Context, biz, phone, code string) error {
	res, err := cc.client.Eval(c, luaSetCode, []string{cc.key(biz, phone)}, code).Int()
	if err != nil {
		return fmt.Errorf("error. %w\n", err)
	}
	// 根据lua脚本来写情况
	switch res {
	case 0:
		return nil
	case -1:
		// 要注意 phone 不能直接记，要脱敏123xxxx286，或加密
		// 要在对应的告警系统里配置，比如说规则，一分钟内出现100次 Warn, 则告警
		//zap.L().Warn("code send too many", zap.String("biz", biz), zap.String("phone", phone))
		return fmt.Errorf("code send too many. %w\n", err)
	default:
		// 系统错误，比如说 -2，是 key 冲突
		return fmt.Errorf("unknown for code. %w\n", err)
	}
}

// Verify 验证验证码
// 如果验证码是一致的，那么删除
// 如果验证码不一致，那么保留的
func (cc *RedisCodeCache) Verify(c context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := cc.client.Eval(c, luaVerifyCode, []string{cc.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, fmt.Errorf("error. %w\n", err)
	}
	// 根据lua脚本来写情况
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, fmt.Errorf("code varify too many. %w\n", err)
	default:
		return false, fmt.Errorf("code error. %w\n", err)
	}
}

func (cc *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code: %s:%s", biz, phone)
}

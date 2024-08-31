package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"webook/internal/domain"
)

type UserCache interface {
	Get(c context.Context, id int64) (domain.User, error)
	Set(c context.Context, u domain.User) error
	Delete(ctx context.Context, id int64) error
}

type RedisUserCache struct {
	// 此方法可传单机的 redis 或 cluster 的 redis
	cmd        redis.Cmdable
	expiration time.Duration
}

// NewUserCache A 用到了 B, B 一定是接口 => 保证面向接口
// A 用到了 B, B 一定是 A 的字段 => 规避宝变量、包方法，都非常缺乏扩展性
// A 用到了 B, A 绝对不初始化 B, 而是外部注入 => 保持依赖注入(DI, Dependency Injection)
func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

// Get 只要 error 为 nil, 就认为缓存里有数据
func (cache *RedisUserCache) Get(c context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.cmd.Get(c, key).Bytes()
	if err != nil {
		return domain.User{}, fmt.Errorf("error. %w\n", err)
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	if err != nil {
		return domain.User{}, fmt.Errorf("error. %w\n", err)
	}
	return u, nil
}

func (cache *RedisUserCache) Set(c context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.cmd.Set(c, key, val, cache.expiration).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user: info: %d\n", id)
}

func (cache *RedisUserCache) Delete(c context.Context, id int64) error {
	return cache.cmd.Del(c, cache.key(id)).Err()
}

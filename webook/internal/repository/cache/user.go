package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"webook/internal/domain"
)

type UserCache struct {
	// 此方法可传单机的 redis 或 cluster 的 redis
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserCache(client redis.Cmdable) *UserCache {
	// A 用到了 B, B 一定是接口
	// A 用到了 B, B 一定是 A 的字段
	// A 用到了 B, A 绝对不初始化 B, 而是外部注入
	return &UserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

// Get 只要 error 为 nil, 就认为缓存里有数据
func (cache *UserCache) Get(c context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(c, key).Bytes()
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

func (cache *UserCache) Set(c context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(c, key, val, cache.expiration).Err()
}

func (cache *UserCache) key(id int64) string {
	return fmt.Sprintf("user: info: %d\n", id)
}

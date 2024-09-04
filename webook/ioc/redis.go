package ioc

import (
	"github.com/redis/go-redis/v9"
	"webook/pkg/cfg"
)

func InitRedis(c cfg.Config) redis.Cmdable {
	r := c.Redis
	cmd := redis.NewClient(&redis.Options{
		Addr: r.Addr,
		// Password: r.Password,
		// DB:       r.DB,
	})
	return cmd
}

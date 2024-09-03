package ratelimit

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"webook/pkg/ratelimit"

	"github.com/gin-gonic/gin"
)

type Builder struct {
	prefix  string
	limiter ratelimit.Limiter
}

func NewBuilder(limiter ratelimit.Limiter) *Builder {
	return &Builder{
		prefix:  "ip-limiter",
		limiter: limiter,
	}
}

func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

// Build 基于redis对ip对限流
func (b *Builder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		limited, err := b.limit(c)
		if err != nil {
			log.Println(err)
			// 这一步很有意思，就是如果这边出错了
			// 要怎么办？
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			log.Println(err)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}

func (b *Builder) limit(c *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", b.prefix, c.ClientIP())
	return b.limiter.Limit(c, key)
}

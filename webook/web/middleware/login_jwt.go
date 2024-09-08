package middleware

import (
	"encoding/gob"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
	ijwt "webook/web/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
	cmd   redis.Cmdable
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(jwtHdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: jwtHdl,
	}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	// 用 Go 的方式编码解码
	gob.Register(time.Now())
	return func(c *gin.Context) {
		// 不需要登录校验
		for _, path := range l.paths {
			if c.Request.URL.Path == path {
				return
			}
		}

		tokenStr := l.ExtractToken(c)
		claims := &ijwt.UserClaims{}
		// ParseWithClaims 里一定要传入 claims 指针，会被解析出来
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), nil
		})
		if err != nil {
			// 没登录
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 情况3: err为nil, token不为nil
		if token == nil || !token.Valid || claims.Uid == 0 {
			// 没登录
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims.UserAgent != c.Request.UserAgent() {
			// 出现严重的安全问题 要监控
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		err = l.CheckSession(c, claims.Ssid)
		if err != nil {
			// 要么 redis 有问题，要么已经退出了登录
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("claims", claims)
	}
}

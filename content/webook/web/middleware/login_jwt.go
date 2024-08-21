package middleware

import (
	"encoding/gob"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
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
		// 用 JWT 来登录校验
		tokenHeader := c.GetHeader("Authorization")

		// 情况1：没带token
		if tokenHeader == "" {
			// 没登录
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 情况2: 带了token，但格式不对
		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 情况3: 带了token,但内容有误
		tokenStr := segs[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), nil
		})
		if err != nil {
			// 没登录
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 情况4: err为nil, token不为nil
		if token == nil || !token.Valid {
			// 没登录
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 刷新
		// expireTime, err :=
	}
}

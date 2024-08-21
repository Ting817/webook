package middleware

import (
	"encoding/gob"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"junior-engineer-training/webook/web"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
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
		tokenHeader := c.GetHeader("authorization")

		// 情况1：没带token
		if tokenHeader == "" {
			// 没登录
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 情况2: 带了token，但格式/内容不对
		// segs := strings.SplitN(tokenHeader, " ", 2)
		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		claims := &web.UserClaims{}
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
		// 登录刷新 每10s刷新一次 (jwt有效期1min)
		now := time.Now()
		if claims.ExpiresAt.Sub(now) < time.Second*50 {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			tokenStr, err = token.SignedString([]byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"))
			if err != nil {
				log.Println("jwt 续约失败", err)
			}
			c.Header("x-jwt-token", tokenStr)
		}

		c.Set("claims", claims)
	}
}

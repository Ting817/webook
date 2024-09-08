package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	SetLoginToken(c *gin.Context, uid int64) error
	SetJWTToken(c *gin.Context, uid int64, ssid string) error
	ClearToken(c *gin.Context) error
	CheckSession(c *gin.Context, ssid string) error
	ExtractToken(c *gin.Context) string
}

type RefreshClaims struct {
	Uid  int64
	Ssid string
	jwt.RegisteredClaims
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64 // Uid: 额外加自己的数据在token里
	UserAgent string
	Ssid      string
}

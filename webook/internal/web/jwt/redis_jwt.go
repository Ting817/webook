package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
	"time"
)

var (
	AtKey = []byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK") // access_token key
	RtKey = []byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPJ") // refresh_token key
)

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		cmd: cmd,
	}
}

func (j *RedisJWTHandler) SetLoginToken(c *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := j.SetJWTToken(c, uid, ssid)
	if err != nil {
		return fmt.Errorf("error %w\n", err)
	}
	err = j.setRefreshToken(c, uid, ssid)
	return err
}

func (j *RedisJWTHandler) ClearToken(c *gin.Context) error {
	c.Header("x-jwt-token", "")
	c.Header("x-refresh-token", "")
	claims := c.MustGet("user").(UserClaims)
	return j.cmd.Set(c, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", time.Hour*24*7).Err()
}

func (j *RedisJWTHandler) CheckSession(c *gin.Context, ssid string) error {
	val, err := j.cmd.Exists(c, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch err {
	case redis.Nil:
		return nil
	case nil:
		if val == 0 {
			return nil
		}
		return errors.New("session is invalid")
	default:
		return err
	}
}

func (j *RedisJWTHandler) ExtractToken(c *gin.Context) string {
	// 用 JWT 来登录校验
	tokenHeader := c.GetHeader("Authorization")
	// segs := strings.SplitN(tokenHeader, " ", 2)
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

func (j *RedisJWTHandler) SetJWTToken(c *gin.Context, uid int64, ssid string) error {
	// 步骤2 在此设置JWT登录态 生成一个JWT token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid: uid,
		// UserAgent: c.GetHeader("User-Agent"),
		UserAgent: c.Request.UserAgent(),
		Ssid:      ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(AtKey)
	if err != nil {
		c.String(http.StatusInternalServerError, "system error,"+err.Error())
		return err
	}
	c.Header("x-jwt-token", tokenStr)
	return nil
}

func (j *RedisJWTHandler) setRefreshToken(c *gin.Context, uid int64, ssid string) error {
	// 步骤2 在此设置JWT登录态 生成一个JWT token
	claims := RefreshClaims{
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid: uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(RtKey)
	if err != nil {
		c.String(http.StatusInternalServerError, "system error,"+err.Error())
		return err
	}
	c.Header("x-refresh-token", tokenStr)
	return nil
}

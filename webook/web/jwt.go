package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type jwtHandler struct {
}

func (j jwtHandler) setJWTToken(c *gin.Context, uid int64) error {
	// 步骤2 在此设置JWT登录态 生成一个JWT token
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid: uid,
		// UserAgent: c.GetHeader("User-Agent"),
		UserAgent: c.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"))
	if err != nil {
		c.String(http.StatusInternalServerError, "system error,"+err.Error())
		return err
	}
	c.Header("x-jwt-token", tokenStr)
	return nil
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64 // Uid: 额外加自己的数据在token里
	UserAgent string
}

package ginx

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64 // Uid: 额外加自己的数据在token里
	UserAgent string
	Ssid      string
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

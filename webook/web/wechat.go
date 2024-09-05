package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"time"
	"webook/internal/service"
	"webook/internal/service/oauth2/wechat"
)

type StateClaims struct {
	State string
	jwt.RegisteredClaims
}

type WechatHandlerConfig struct {
	Secure bool
}

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	jwtHandler
	stateKey []byte
	cfg      WechatHandlerConfig
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, cfg WechatHandlerConfig) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:      svc,
		userSvc:  userSvc,
		stateKey: []byte("Cb3cErlIjTEzxHwr6uhsMZ8On5s5EMPK"),
		cfg:      cfg,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(c *gin.Context) {
	state := uuid.New()
	url, err := h.svc.AuthURL(c, state)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造扫码登录失败",
		})
		return
	}

	if err = h.setStateCookie(c, state); err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}

	c.JSON(http.StatusOK, Result{
		Date: url,
	})
}

func (h *OAuth2WechatHandler) setStateCookie(c *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return fmt.Errorf("error %w\n" + err.Error())
	}
	// secure 和 httpOnly 同时设为 true，则攻击者不会拿到 cookie // 在 ioc 控制 secure，外面传进来
	c.SetCookie("jwt-state", tokenStr, 600, "/oauth2/wechat/callback", "", h.cfg.Secure, true)
	return nil
}

func (h *OAuth2WechatHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	err := h.verifyState(c)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}
	info, err := h.svc.VerifyCode(c, code)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}

	u, err := h.userSvc.FindOrCreateByWechat(c, info)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}
	err = h.setJWTToken(c, u.Id)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "system error",
		})
		return
	}
	c.JSON(http.StatusOK, Result{
		Msg: "come here!",
	}) // 确认拿到 wechat 的 code
}

func (h *OAuth2WechatHandler) verifyState(c *gin.Context) error {
	state := c.Query("state")
	// 检验一下 state
	ck, err := c.Cookie("jwt-state")
	if err != nil {
		return fmt.Errorf("get jwt-state failed. %w/n" + err.Error())
	}
	var sc StateClaims
	token, err := jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("token already is invalid. %w\n", err)
	}

	if sc.State != state {
		return fmt.Errorf("state wrong")
	}
	return nil
}

package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/internal/service"
	"webook/internal/service/oauth2/wechat"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	jwtHandler
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(c *gin.Context) {
	url, err := h.svc.AuthURL(c)
	if err != nil {
		c.JSON(http.StatusOK, Result{
			Date: url,
		})
	}
}

func (h *OAuth2WechatHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	info, err := h.svc.VerifyCode(c, code, state)
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

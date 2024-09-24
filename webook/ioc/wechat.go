package ioc

import (
	"os"
	"webook/internal/service/oauth2/wechat"
	"webook/internal/web"
	"webook/pkg/logger"
)

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("Not found env WECHAT_APP_ID")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("Not found env WECHAT_APP_SECRET")
	}
	var l logger.LoggerV1
	return wechat.NewService(appId, appKey, l)
}

func NewWechatHandlerConfig() web.WechatHandlerConfig {
	return web.WechatHandlerConfig{
		Secure: false,
	}
}

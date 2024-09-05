//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/ioc"
	"webook/web"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 基础部分
		ioc.NewCfg, ioc.InitDB, ioc.InitRedis,

		// DAO 部分
		dao.NewUserDAO,

		// Cache 部分
		cache.NewUserCache, cache.NewCodeCache,

		// repository 部分
		repository.NewUserRepository, repository.NewCodeRepository,

		// service 部分
		ioc.InitSmsService, ioc.InitWechatService, service.NewUserService, service.NewSMSCodeService,

		// handler 部分
		web.NewUserHandler, web.NewOAuth2WechatHandler, ioc.NewWechatHandlerConfig,

		// gin 的中间件
		ioc.InitMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	return gin.Default()
}

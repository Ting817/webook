//go:build wireinject

package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/article"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	article2 "webook/internal/repository/dao/article"
	"webook/internal/service"
	web2 "webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 基础部分
		ioc.NewCfg,
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitLogger,

		// DAO 部分
		dao.NewUserDAO,
		article2.NewGORMArticleDAO,

		// Cache 部分
		cache.NewUserCache,
		cache.NewCodeCache,

		// repository 部分
		repository.NewUserRepository,
		repository.NewCodeRepository,
		article.NewArticleRepository,

		// service 部分
		ioc.InitSmsService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewSMSCodeService,
		service.NewArticleService,

		// handler 部分
		web2.NewUserHandler,
		web2.NewOAuth2WechatHandler,
		ioc.NewWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,
		web2.NewArticleHandler,

		// gin 的中间件
		ioc.InitMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	return gin.Default()
}

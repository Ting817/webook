//go:build wireinject

package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	article2 "webook/internal/repository/article"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/repository/dao/article"
	"webook/internal/service"
	"webook/internal/web"
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
		article.NewGORMArticleDAO,

		// Cache 部分
		cache.NewRedisUserCache,
		cache.NewRedisCodeCache,
		cache.NewRedisArticleCache,

		// repository 部分
		repository.NewUserRepository,
		repository.NewCodeRepository,
		article2.NewArticleRepository,

		// service 部分
		ioc.InitSmsService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewSMSCodeService,
		service.NewArticleService,

		// handler 部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.NewWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,
		web.NewArticleHandler,

		// gin 的中间件
		ioc.InitMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	return gin.Default()
}

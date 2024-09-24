//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/article"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	web2 "webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"
)

var thirdProvider = wire.NewSet(InitRedis, InitTestDB, InitLog)
var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)
var articleSvcProvider = wire.NewSet(
	dao.NewGORMArticleDAO,
	article.NewArticleRepository,
	service.NewArticleService)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 基础部分
		thirdProvider,
		userSvcProvider,
		articleSvcProvider,

		// 验证码缓存在redis中
		cache.NewCodeCache,
		repository.NewCodeRepository,

		// service 部分
		ioc.InitSmsService, service.NewSMSCodeService,

		// 指定啥也不干的 wechat service
		InitPhantomWechatService,

		// handler 部分
		web2.NewUserHandler, web2.NewArticleHandler, web2.NewOAuth2WechatHandler, ioc.NewWechatHandlerConfig, ijwt.NewRedisJWTHandler,

		// gin 的中间件
		ioc.InitMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler() *web2.ArticleHandler {
	wire.Build(thirdProvider,
		dao.NewGORMArticleDAO,
		service.NewArticleService,
		web2.NewArticleHandler,
		article.NewArticleRepository,
	)
	return &web2.ArticleHandler{}
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}

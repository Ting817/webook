//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/article"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	article2 "webook/internal/repository/dao/article"
	"webook/internal/service"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"
)

var thirdProvider = wire.NewSet(InitRedis, InitTestDB, InitLog)
var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewRedisUserCache,
	repository.NewUserRepository,
	service.NewUserService)
var articleSvcProvider = wire.NewSet(
	article2.NewGORMArticleDAO,
	cache.NewRedisArticleCache,
	article.NewArticleRepository,
	service.NewArticleService)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 基础部分
		thirdProvider,
		userSvcProvider,
		articleSvcProvider,
		interactiveSvcProvider,

		// 验证码缓存在redis中
		cache.NewRedisCodeCache,
		repository.NewCodeRepository,

		// service 部分
		ioc.InitSmsService, service.NewSMSCodeService,

		// 指定啥也不干的 wechat service
		InitPhantomWechatService,

		// handler 部分
		web.NewUserHandler, web.NewArticleHandler, web.NewOAuth2WechatHandler, ioc.NewWechatHandlerConfig, ijwt.NewRedisJWTHandler,

		// gin 的中间件
		ioc.InitMiddlewares,

		// Web 服务器
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(dao article2.ArticleDAO) *web.ArticleHandler {
	wire.Build(thirdProvider,
		//article2.NewGORMArticleDAO,
		userSvcProvider,
		interactiveSvcProvider,
		cache.NewRedisArticleCache,
		service.NewArticleService,
		web.NewArticleHandler,
		article.NewArticleRepository,
	)
	return new(web.ArticleHandler)
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service.NewInteractiveService(nil, nil)
}

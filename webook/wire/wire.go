//go:build wireinject

package wire

import (
	"github.com/google/wire"
	article3 "webook/internal/events/article"
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

func InitApp() *App {
	wire.Build(
		// 基础部分
		ioc.NewCfg,
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,

		// events 部分
		article3.NewKafkaConsumer,
		article3.NewKafkaProducer,

		// DAO 部分
		dao.NewUserDAO,
		dao.NewGORMInteractiveDAO,
		article.NewGORMArticleDAO,

		// Cache 部分
		cache.NewRedisUserCache,
		cache.NewRedisCodeCache,
		cache.NewRedisArticleCache,
		cache.NewRedisInteractiveCache,

		// repository 部分
		repository.NewUserRepository,
		repository.NewCodeRepository,
		article2.NewArticleRepository,
		repository.NewCachedInteractiveRepository,

		// service 部分
		ioc.InitSmsService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewSMSCodeService,
		service.NewArticleService,
		service.NewInteractiveService,

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
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

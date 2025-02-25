package ioc

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	web2 "webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/internal/web/middleware"
	"webook/pkg/ginx"
	"webook/pkg/logger"
	"webook/pkg/middlewares/accesslog"
)

func InitWebServer(mdls []gin.HandlerFunc, hdl *web2.UserHandler, oauth2WechatHdl *web2.OAuth2WechatHandler, articleHdl *web2.ArticleHandler, l logger.LoggerV1) *gin.Engine {
	ginx.SetLogger(l)
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler, l logger.LoggerV1) []gin.HandlerFunc {

	bd := accesslog.NewMiddlewareBuilder(func(c context.Context, al *accesslog.AccessLog) {
		l.Debug("Gin 收到请求", logger.Field{
			Key:   "req",
			Value: al,
		})
	}).AllowReqBody(true).AllowRespBody()

	return []gin.HandlerFunc{
		corsHdl(),
		//ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
		bd.Build(),
		// 使用 JWT
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login").
			IgnorePaths("/users/refresh_token").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			Build(),

		// 使用session 登录校验
		// sessionHandlerFunc(),
		// middleware.NewLoginMiddlewareBuilder().
		//		IgnorePaths("/users/signup").
		//		IgnorePaths("/users/login").Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		// AllowOrigins: []string{"http://localhost:3000"},
		AllowPrivateNetwork: true,
		AllowHeaders:        []string{"Content-Type", "Authorization"},
		ExposeHeaders:       []string{"x-jwt-token", "x-refresh-token"}, // 给前端token
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, ":3000") // 通常是公司的域名
		},
		MaxAge: 12 * time.Hour,
	})
}

func sessionHandlerFunc() gin.HandlerFunc {
	// store := cookie.NewStore([]byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), []byte("Hg2WjnYiGz4XUNVhBUNAIrSu35Z7uyPA"))

	store := memstore.NewStore([]byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), []byte("Hg2WjnYiGz4XUNVhBUNAIrSu35Z7uyPA"))

	// store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), []byte("Hg2WjnYiGz4XUNVhBUNAIrSu35Z7uyPA"))
	// if err != nil {
	// 	panic(err)
	// }

	// 实现sqlx
	// myStore := sqlx_store.Store{}

	// cookie的名字叫做ssid
	return sessions.Sessions("ssid", store)
}

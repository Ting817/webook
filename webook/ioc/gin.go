package ioc

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"webook/pkg/ginx/middlewares/ratelimit"
	"webook/web"
	"webook/web/middleware"
)

func InitWebServer(mdls []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),

		// 使用 JWT
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").Build(),

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
		ExposeHeaders:       []string{"x-jwt-token"}, // 给前端token
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

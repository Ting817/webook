package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	ijwt "webook/internal/web/jwt"
	middleware2 "webook/internal/web/middleware"
)

func initWebServer(redisCmd redis.Cmdable, jwtHdl ijwt.Handler) *gin.Engine {
	server := gin.Default()
	//server.Use(ratelimit.NewBuilder(redisCmd, time.Minute, 100).Build())
	server.Use(corsHandler())
	// 使用 session 机制登录
	// usingSession(server)
	// 使用 JWT
	usingJWT(server, jwtHdl)
	// 注册路由
	return server
}

func corsHandler() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowCredentials: true,
		// 在使用 JWT 的时候，因为我们使用了 Authorization 的头部，所以要加上
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 为了 JWT
		ExposeHeaders: []string{"X-Jwt-Token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	})
}

func usingJWT(server *gin.Engine, jwtHdl ijwt.Handler) {
	mldBd := middleware2.NewLoginJWTMiddlewareBuilder(jwtHdl)
	server.Use(mldBd.Build())
}

func usingSession(server *gin.Engine) {
	// store := cookie.NewStore([]byte("secret"))

	// 这是基于内存的实现，第一个参数是 authentication key，最好是 32 或者 64 位
	// 第二个参数是 encryption key
	store := memstore.NewStore([]byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"),
		[]byte("o6jdlo2cb9f9pb6h46fjmllw481ldebj"))
	// 第一个参数是最大空闲连接数量
	// 第二个就是 tcp，你不太可能用 udp
	// 第三、四个 就是连接信息和密码
	// 第五第六就是两个 key
	// store, err := redis.NewStore(16, "tcp",
	//	"localhost:6379", "",
	//	// authentication key, encryption key
	//	[]byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"),
	//	[]byte("o6jdlo2cb9f9pb6h46fjmllw481ldebj"))
	// if err != nil {
	//	panic(err)
	// }

	// cookie 的名字叫做ssid
	server.Use(sessions.Sessions("ssid", store))
	// 登录校验
	login := middleware2.NewLoginMiddlewareBuilder()
	server.Use(login.Build())
}

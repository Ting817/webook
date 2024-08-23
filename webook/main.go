package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	// "github.com/gin-contrib/sessions/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/pkg/ginx/middlewares/ratelimit"
	"webook/web"
	"webook/web/middleware"
)

func main() {
	// db := initDB()
	// u := initUser(db)
	// // server := initWebServer()  // 用session
	// server := initWebServerJWT() // 用JWT
	// u.RegisterRoutes(server)
	server := gin.Default()
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello, welcome to here")
	})
	server.Run(":8080")
}

func initWebServerJWT() *gin.Engine {
	server := gin.Default()
	// 跨域问题 Use()会作用于此r的全部的路由
	server.Use(cors.New(cors.Config{
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
		MaxAge: 60 * time.Second,
	}))

	// 实现登录功能
	// 步骤3 用JWT
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").Build())

	// 限流
	cmd := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})
	// 1min 100次
	server.Use(ratelimit.NewBuilder(cmd, time.Minute, 100).Build())

	return server
}

// func initWebServer() *gin.Engine {
// 	server := gin.Default()
// 	// 跨域问题 Use()会作用于此r的全部的路由
// 	server.Use(cors.New(cors.Config{
// 		// AllowOrigins: []string{"http://localhost:3000"},
// 		AllowPrivateNetwork: true,
// 		AllowHeaders:        []string{"content-Type", "authorization"},
// 		AllowCredentials:    true, // 允许带cookie之类的
// 		AllowOriginFunc: func(origin string) bool {
// 			if strings.HasPrefix(origin, "http://localhost") {
// 				return true
// 			}
// 			return strings.Contains(origin, ":3000") // 通常是公司的域名
// 		},
// 		MaxAge: 12 * time.Hour,
// 	}))
//
// 	// 实现登录功能 步骤1
// 	// store := cookie.NewStore([]byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), []byte("Hg2WjnYiGz4XUNVhBUNAIrSu35Z7uyPA"))
//
// 	store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), []byte("Hg2WjnYiGz4XUNVhBUNAIrSu35Z7uyPA"))
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// 实现sqlx
// 	// myStore := sqlx_store.Store{}
// 	// server.Use(sessions.Sessions("mysession", myStore))
//
// 	server.Use(sessions.Sessions("mysession", store)) // cookie的名字叫做mysession
//
// 	// 步骤3 用session
// 	server.Use(middleware.NewLoginMiddlewareBuilder().
// 		IgnorePaths("/users/signup").
// 		IgnorePaths("/users/login").Build())
//
// 	return server
// }

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		panic(err) // panic 即goroutine直接结束 只在初始化时考虑panic
	}
	if err = dao.InitTable(db); err != nil {
		panic(err)
	}

	return db
}

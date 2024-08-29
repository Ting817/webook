package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/pelletier/go-toml/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"

	// "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/service/sms/memory"
	"webook/pkg/ginx/middlewares/ratelimit"
	"webook/web"
	"webook/web/middleware"
)

type Config struct {
	K8s struct {
		Addr      string `toml:"addr"`
		Token     string `toml:"token"`
		Namespace string `toml:"namespace"`
	} `toml:"k8s"`
	DB struct {
		DSN string `toml:"dsn"`
	} `toml:"db"`
	Redis struct {
		Addr string `toml:"addr"`
	} `toml:"redis"`
}

func main() {
	db := initDB()
	redisCmd := initRedis()
	u := initUser(db, redisCmd)
	// server := initWebServer() // 用session
	server := initWebServerJWT() // 用JWT
	u.RegisterRoutes(server)
	// server := gin.Default()
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
		IgnorePaths("/users/login").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms").Build())

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

func initWebServer() *gin.Engine {
	server := gin.Default()

	configFile := "config/default.toml"
	fileContent, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error decoding TOML file: %v", err)
	}
	var config Config
	if err = toml.Unmarshal(fileContent, &config); err != nil {
		log.Fatalf("error decoding toml file: %v", err)
	}

	// 限流
	// redisClient := redis.NewClient(&redis.Options{
	// 	Addr: config.Redis.Addr,
	// })
	// server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	// 跨域问题 Use()会作用于此r的全部的路由
	server.Use(cors.New(cors.Config{
		// AllowOrigins: []string{"http://localhost:3000"},
		AllowPrivateNetwork: true,
		AllowHeaders:        []string{"content-Type", "authorization"},
		AllowCredentials:    true, // 允许带cookie之类的
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, ":3000") // 通常是公司的域名
		},
		MaxAge: 12 * time.Hour,
	}))

	// 实现登录功能 步骤1
	// store := cookie.NewStore([]byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), []byte("Hg2WjnYiGz4XUNVhBUNAIrSu35Z7uyPA"))

	store := memstore.NewStore([]byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), []byte("Hg2WjnYiGz4XUNVhBUNAIrSu35Z7uyPA"))

	// store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("Cb3cErlIjTEzfHwr6uhsMZ8On5s5EMPK"), []byte("Hg2WjnYiGz4XUNVhBUNAIrSu35Z7uyPA"))
	// if err != nil {
	// 	panic(err)
	// }

	// 实现sqlx
	// myStore := sqlx_store.Store{}
	// server.Use(sessions.Sessions("mysession", myStore))

	server.Use(sessions.Sessions("mysession", store)) // cookie的名字叫做mysession

	// 步骤3 用session
	server.Use(middleware.NewLoginMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").Build())

	return server
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	c := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(ud, c)
	svc := service.NewUserService(repo)
	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(smsSvc, codeRepo)
	u := web.NewUserHandler(svc, codeSvc)
	return u
}

func initDB() *gorm.DB {
	configFile := "config/default.toml"
	fileContent, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error decoding TOML file: %v", err)
	}
	var config Config

	if err = toml.Unmarshal(fileContent, &config); err != nil {
		log.Fatalf("error decoding toml file: %v", err)
	}
	fmt.Print("config---", config)

	db, err := gorm.Open(mysql.Open(config.DB.DSN), &gorm.Config{})
	if err != nil {
		panic(err) // panic 即goroutine直接结束 只在初始化时考虑panic
	}
	if err = dao.InitTable(db); err != nil {
		panic(err)
	}

	return db
}

func initRedis() redis.Cmdable {
	configFile := "config/default.toml"
	fileContent, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error decoding TOML file: %v", err)
	}
	var config Config
	if err = toml.Unmarshal(fileContent, &config); err != nil {
		log.Fatalf("error decoding toml file: %v", err)
	}
	rCfg := config.Redis
	cmd := redis.NewClient(&redis.Options{
		Addr: rCfg.Addr,
		// Password: rCfg.Password,
		// DB:       rCfg.DB,
	})
	return cmd
}

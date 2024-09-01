package ioc

import (
	"fmt"
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"webook/internal/repository/dao"
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

func InitDB() *gorm.DB {
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

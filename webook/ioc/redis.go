package ioc

import (
	"log"
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	configFile := "config/default.toml"
	fileContent, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error decoding TOML file: %v", err)
	}
	var config Config
	if err = toml.Unmarshal(fileContent, &config); err != nil {
		log.Fatalf("error decoding toml file: %v", err)
	}
	r := config.Redis
	cmd := redis.NewClient(&redis.Options{
		Addr: r.Addr,
		// Password: r.Password,
		// DB:       r.DB,
	})
	return cmd
}

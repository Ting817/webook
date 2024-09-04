package ioc

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"log"
	"os"
	"webook/pkg/cfg"
)

func NewCfg() cfg.Config {
	configFile := "config/default.toml"
	fileContent, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error decoding TOML file: %v", err)
	}
	var config cfg.Config

	if err = toml.Unmarshal(fileContent, &config); err != nil {
		log.Fatalf("error decoding toml file: %v", err)
	}
	fmt.Print("config---", config)

	return config
}

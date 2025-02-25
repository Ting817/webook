package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"webook/internal/events"
	"webook/internal/events/article"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `toml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

func NewConsumers(c1 article.Consumer) []events.Consumer {
	return []events.Consumer{c1}
}

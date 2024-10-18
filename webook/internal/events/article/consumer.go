package article

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/internal/repository"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

const topicReadEvent = "article_read_event"

type Consumer interface {
	Start() error
}

type KafkaConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.LoggerV1
}

func NewKafkaConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.LoggerV1) Consumer {
	return &KafkaConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (k *KafkaConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", k.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{topicReadEvent}, saramax.NewHandler[ReadEvent](k.l, k.Consume))
		if err != nil {
			k.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (k *KafkaConsumer) Consume(msg *sarama.ConsumerMessage, evt ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := k.repo.IncrReadCnt(ctx, "article", evt.Aid)
	return err
}

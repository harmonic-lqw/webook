package events

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/interactive/repository"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type BatchInteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	l      logger.LoggerV1
}

func NewBatchInteractiveReadEventConsumer(repo repository.InteractiveRepository, client sarama.Client, l logger.LoggerV1) *BatchInteractiveReadEventConsumer {
	return &BatchInteractiveReadEventConsumer{repo: repo, client: client, l: l}
}

func (i *BatchInteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewBatchHandler[ReadEvent](i.BatchConsume, i.l))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (i *BatchInteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, evt.Aid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	return i.repo.BatchIncrReadCnt(ctx, bizs, bizIds)
}

package article

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type HistoryRecordConsumer struct {
	repo   repository.HistoryRecordRepository
	client sarama.Client
	l      logger.LoggerV1
}

func (h *HistoryRecordConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", h.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewBatchHandler[ReadEvent](h.BatchConsume, h.l))
		if er != nil {
			h.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (h *HistoryRecordConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))
	var his []domain.HistoryRecord
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, evt.Aid)
		his = append(his, domain.HistoryRecord{
			BizId: evt.Aid,
			Biz:   "article",
			Uid:   evt.Uid,
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()

	return h.repo.BatchAddRecord(ctx, his)
}

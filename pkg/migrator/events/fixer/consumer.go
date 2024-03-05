package fixer

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
	"webook/pkg/migrator/fixer"
	"webook/pkg/saramax"
)

type FixConsumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.LoggerV1
	srcFirst *fixer.Fixer[T]
	dstFirst *fixer.Fixer[T]
	topic    string
}

func NewFixConsumer[T migrator.Entity](client sarama.Client, l logger.LoggerV1, src *gorm.DB, dst *gorm.DB, topic string) (*FixConsumer[T], error) {
	srcFirst, err := fixer.NewFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	return &FixConsumer[T]{client: client,
		l:        l,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
		topic:    topic}, nil
}

func (f *FixConsumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("fix", f.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{f.topic}, saramax.NewHandler[events.InconsistentEvent](f.Consume, f.l))
		if er != nil {
			f.l.Error("数据修复消费者异常", logger.Error(er))
		}
	}()
	return nil
}

func (f *FixConsumer[T]) Consume(msg *sarama.ConsumerMessage, event events.InconsistentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	switch event.Direction {
	case "SRC":
		return f.srcFirst.Fix(ctx, event.ID)
	case "DST":
		return f.dstFirst.Fix(ctx, event.ID)
	}
	return errors.New("未知的校验方向")
}

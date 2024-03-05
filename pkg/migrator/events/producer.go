package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error
}

type SaramaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewSaramaProducer(producer sarama.SyncProducer, topic string) Producer {
	return &SaramaProducer{producer: producer, topic: topic}
}

func (s *SaramaProducer) ProduceInconsistentEvent(ctx context.Context, evt InconsistentEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(val),
	})
	return err
}

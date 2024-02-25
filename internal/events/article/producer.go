package article

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const TopicReadEvent = "article_read"

type Producer interface {
	ProduceReadEvent(evt ReadEvent) error
}

type ReadEvent struct {
	Aid int64
	Uid int64
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
	vector   *prometheus.SummaryVec
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "harmonic",
		Subsystem: "webook",
		Name:      "kafka" + "_produce_time",
		Help:      "统计向 kafka 发送数据的时间",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"topic"})
	prometheus.MustRegister(vector)
	return &SaramaSyncProducer{producer: producer,
		vector: vector,
	}
}

func (s *SaramaSyncProducer) ProduceReadEvent(evt ReadEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	start := time.Now()
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicReadEvent,
		Value: sarama.StringEncoder(val),
	})
	duration := time.Since(start).Milliseconds()
	s.vector.WithLabelValues(TopicReadEvent).Observe(float64(duration))
	return err
}

package sarama

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var addr = []string{"localhost:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(addr, cfg)
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "test_topic",
			Value: sarama.StringEncoder("一条同步消息"),
			// 类似于 http 里的 header，会在生产者和消费者之间传递
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("key1"),
					Value: []byte("value1"),
				},
			},
			Metadata: "这是 metadata",
		})
	}

}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addr, cfg)
	assert.NoError(t, err)
	msgs := producer.Input()
	msgs <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("一条异步消息"),
		// 类似于 http 里的 header，会在生产者和消费者之间传递
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("key1"),
				Value: []byte("value1"),
			},
		},
		Metadata: "这是 metadata",
	}

	select {
	case msg := <-producer.Successes():
		t.Log("发送成功", msg.Value.(sarama.StringEncoder))
	case err := <-producer.Errors():
		t.Log("发送失败", err.Err, err.Msg)
	}
}

func TestSelect(t *testing.T) {
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)
	go func() {
		ch1 <- 123
		time.Sleep(time.Second)
	}()
	go func() {
		ch2 <- 234
		time.Sleep(time.Second)
	}()
	select {
	case val := <-ch1:
		t.Log("casech1: ", val)
	case val := <-ch2:
		t.Log("casech2: ", val)
	}

	t.Log("after select ch1", <-ch1)
	t.Log("after select ch2", <-ch2)
}

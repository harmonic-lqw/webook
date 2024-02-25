package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addr, "demo", cfg)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	// 一直阻塞到这里消费，直到超时或上下文被取消
	err = consumer.Consume(ctx, []string{"test_topic"}, ConsumerHandler{})
	assert.NoError(t, err)
}

type ConsumerHandler struct {
}

func (c ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("这是 Setup")
	// 可以重置偏移量 offset
	//partitions := session.Claims()["test_topic"]
	//var offset int64 = 0
	//for _, part := range partitions {
	//	session.ResetOffset("test_topic", part, offset, "")
	//}
	return nil
}
func (c ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("这是 Cleanup")
	return nil
}

// ConsumeClaim 异步消费 批量提交
func (c ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	const batchSize = 10
	for {
		log.Println("开始一个新的批次")
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		// 为了防止永远凑不够一个 batchSize 导致阻塞，加入 ctx
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var done = false
		// 如果不用 eg，则是 批量消费，批量提交
		var eg errgroup.Group
		for i := 0; i < batchSize; i++ {
			select {
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				batch = append(batch, msg)
				// 异步消费
				eg.Go(func() error {
					log.Println(string(msg.Value))
					return nil
				})
			case <-ctx.Done():
				done = true
			}

			if done {
				break
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			log.Println(err)
			continue
		}

		// 批量提交
		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
}

func (c ConsumerHandler) ConsumeClaimV1(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		log.Println(string(msg.Value))
		// 提交
		session.MarkMessage(msg, "")
	}
	return nil
}

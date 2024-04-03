package filter

import (
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/bloom"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	address := "localhost:6379"

	RedisClient := redis.New(address)
	TestFilter := bloom.New(RedisClient, "BloomKey", 20*300)

	err := TestFilter.Add([]byte("test"))
	assert.NoError(t, err)

	res, err := TestFilter.Exists([]byte("test"))
	assert.NoError(t, err)
	t.Log(res)
}

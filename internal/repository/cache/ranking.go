package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math"
	"strconv"
	"strings"
	"time"
	"webook/internal/domain"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
	SetLoad(ctx context.Context, nodeId int64, load int) error
	GetMinLoadNode(ctx context.Context) (int64, int, error)
}

type RankingRedisCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func NewRankingRedisCache(client redis.Cmdable) RankingCache {
	return &RankingRedisCache{
		client: client,
		key:    "ranking:top_n",
		// 这是 Redis 的过期时间
		expiration: time.Minute * 3}
}

func (r *RankingRedisCache) GetMinLoadNode(ctx context.Context) (int64, int, error) {
	// 节点数预测不会太多，完全遍历应该没问题
	iter := r.client.Scan(ctx, 0, "nodeId:*", 0).Iterator()
	var minNodeId int64 = -1
	var minLoad = math.MaxInt
	var nodeId int64
	var load int
	for iter.Next(ctx) {
		key := iter.Val()
		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			// 出错时返回当前能找到的最小负载节点
			return minNodeId, minLoad, err
		}
		load, err = strconv.Atoi(val)
		if err != nil {
			return minNodeId, minLoad, errors.New("类型转换错误")
		}
		if load < minLoad {
			nodeId, err = strconv.ParseInt(strings.Split(key, ":")[1], 10, 64)
			if err != nil {
				return minNodeId, minLoad, errors.New("类型转换错误")
			}
			minLoad = load
			minNodeId = nodeId
		}
	}
	return minNodeId, minLoad, nil
}

func (r *RankingRedisCache) SetLoad(ctx context.Context, nodeId int64, load int) error {
	nodeKey := r.nodeKey(nodeId)
	return r.client.Set(ctx, nodeKey, load, time.Minute).Err()
}

func (r *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (r *RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := range arts {
		arts[i].Content = arts[i].Abstract()
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key, val, r.expiration).Err()
}

func (r *RankingRedisCache) nodeKey(nodeId int64) string {
	return fmt.Sprintf("nodeId:%d", nodeId)
}

func (r *RankingRedisCache) SetV1(ctx context.Context, score []float64, arts []domain.Article) error {
	if len(score) != len(arts) {
		return errors.New("得分数组和文章数组数量不匹配！")
	}
	pipe := r.client.TxPipeline()
	for i, art := range arts {
		art.Content = art.Abstract()
		val, err := json.Marshal(art)
		if err != nil {
			return err
		}
		pipe.ZAdd(ctx, r.key, redis.Z{
			Score:  score[i],
			Member: val,
		})
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

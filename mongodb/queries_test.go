package mongodb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

type MongoDBTestSuite struct {
	suite.Suite
	col *mongo.Collection
}

func (s *MongoDBTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Println(evt.Command)
		},
	}
	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(s.T(), err)
	// 连接成功，操作 client
	col := client.Database("webook").Collection("articles")
	s.col = col

	manyRes, err := col.InsertMany(ctx, []any{Article{
		Id:       123,
		AuthorId: 11,
	}, Article{
		Id:       234,
		AuthorId: 12,
	},
	})
	assert.NoError(s.T(), err)
	s.T().Log("插入数量", len(manyRes.InsertedIDs))
}

func (s *MongoDBTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := s.col.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.col.Indexes().DropAll(ctx)
	assert.NoError(s.T(), err)
}

func (s *MongoDBTestSuite) TestOr() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	filter := bson.A{bson.D{bson.E{"id", 123}}, bson.D{bson.E{"id", 234}}}
	findRes, err := s.col.Find(ctx, bson.D{bson.E{"$or", filter}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = findRes.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log("查询结果", arts)
}

func (s *MongoDBTestSuite) TestAnd() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	insertMany, err := s.col.InsertMany(ctx, []any{
		Article{
			Id:    345,
			Title: "有标题",
		},
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(insertMany.InsertedIDs))

	filter := bson.A{bson.D{bson.E{"id", 345}}, bson.D{bson.E{"title", "有标题"}}}
	findRes, err := s.col.Find(ctx, bson.D{bson.E{"$and", filter}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = findRes.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log("查询结果", arts)
}

func (s *MongoDBTestSuite) TestIn() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	filter := bson.D{bson.E{"id", // 先给 id
		bson.D{bson.E{"$in", []int{123, 234}}}}}
	findRes, err := s.col.Find(ctx, filter,
		// 只查找特定字段
		options.Find().SetProjection(bson.D{bson.E{"id", 1}}),
	)
	assert.NoError(s.T(), err)
	var arts []Article
	err = findRes.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log("查询结果", arts)
}

func (s *MongoDBTestSuite) TestIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	ires, err := s.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{bson.E{"id", 1}},
		// 设置为唯一索引
		Options: options.Index().SetUnique(true).SetName("idx_id"),
	})
	assert.NoError(s.T(), err)
	s.T().Log("创建索引", ires)
}

func TestMongoDBQueries(t *testing.T) {
	suite.Run(t, &MongoDBTestSuite{})
}

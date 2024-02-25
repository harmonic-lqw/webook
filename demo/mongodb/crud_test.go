package mongodb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestMongoDB(t *testing.T) {
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
	assert.NoError(t, err)
	// 连接成功，操作 client
	col := client.Database("webook").Collection("articles")

	// ######## 插入 ########
	insertRes, err := col.InsertOne(ctx, Article{
		Id:       1, // mongodb 没有自增主键的概念，要手动设置
		Title:    "我的标题",
		Content:  "我的内容",
		AuthorId: 123,
	})
	assert.NoError(t, err)
	oid := insertRes.InsertedID.(primitive.ObjectID)
	t.Log("插入ID ", oid)

	// ######## 查找 ########
	filter := bson.D{bson.E{"id", 1}}
	//filter := bson.M{
	//	"id": 1,
	//}
	findRes := col.FindOne(ctx, filter)
	// 没找到数据
	if findRes.Err() == mongo.ErrNoDocuments {
		t.Log("没找到数据")
	} else {
		assert.NoError(t, findRes.Err())
		var art Article
		err = findRes.Decode(&art)
		assert.NoError(t, err)
		t.Log(art)
	}

	// ######## 更新 ########
	updateFilter := bson.D{bson.E{Key: "id", Value: 1}}
	//set := bson.D{bson.E{Key: "$set", Value: bson.E{Key: "title", Value: "新的标题"}}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title": "新的标题",
	}}}
	updateOneRes, err := col.UpdateOne(ctx, updateFilter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateOneRes.ModifiedCount)

	updateManyRes, err := col.UpdateMany(ctx, updateFilter, bson.D{bson.E{Key: "$set",
		Value: Article{ // 这种写法会导致所有的字段全部更新，未设置的修改为对应的零值
			Content: "新的内容",
		}}})
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)

	// ######## 删除 ########
	deleteFilter := bson.D{bson.E{Key: "id", Value: 1}}
	delRes, err := col.DeleteMany(ctx, deleteFilter)
	assert.NoError(t, err)
	t.Log("删除文档数量", delRes.DeletedCount)
}

type Article struct {
	Id       int64 `bson:"id, omitempty"` // 如此设置，可忽略零值修改
	Title    string
	Content  string
	AuthorId int64 `bson:"author_id, omitempty"`
	Ctime    int64
	Utime    int64
	Status   uint8
}

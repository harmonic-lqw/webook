package dao

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDBArticleDAO struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (m *MongoDBArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	art.Id = m.node.Generate().Int64()
	_, err := m.col.InsertOne(ctx, &art)
	return art.Id, err
}

func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	filter := bson.D{bson.E{"id", art.Id}, bson.E{"author_id", art.AuthorId}}
	set := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("ID 不对或者创作者不对")
	}
	return nil
}

func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	// 制作库 的 Insert Or Update 语义实现
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art) // 这条分支的语义：新建并发表
	}
	if err != nil {
		return 0, err
	}
	// 新建并发表情况下，要先将线上库的 id 赋值
	art.Id = id
	// 线上库 的 Insert Or Update 语义实现
	filter := bson.D{bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId}}
	now := time.Now().UnixMilli()
	pubArt := PublishedArticle(art)
	// 这种更新方式会引起 ctime 的冲突
	//pubArt.Utime = now
	//set := bson.D{bson.E{Key: "$set", Value: pubArt},
	//			bson.E{Key: "$setOnInsert",
	//				Value: bson.D{bson.E{Key: "ctime", Value: now}}}}

	set := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title":   pubArt.Title,
		"content": pubArt.Content,
		"utime":   now,
		"status":  pubArt.Status,
	}},
		bson.E{Key: "$setOnInsert", // 这里就从 dao 层面实现了 Upsert 语义
			Value: bson.D{bson.E{Key: "ctime", Value: now}}}}

	_, err = m.liveCol.UpdateOne(ctx,
		filter, set,
		options.Update().SetUpsert(true))

	return id, err
}

func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	filter := bson.D{bson.E{Key: "id", Value: id},
		bson.E{Key: "author_id", Value: uid}}
	sets := bson.D{bson.E{Key: "$set",
		Value: bson.D{bson.E{Key: "status", Value: status}}}}
	res, err := m.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("ID 不对或者创作者不对")
	}

	_, err = m.liveCol.UpdateOne(ctx, filter, sets)
	return err
}

func NewMongoDBArticleDAO(mdb *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBArticleDAO{
		node:    node,
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
	}
}

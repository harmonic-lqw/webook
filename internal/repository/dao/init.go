package dao

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"time"
)

// InitTables 使用GORM自带的建表功能
// 这是种不太好的做法
func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{},
		&AsyncSms{},
		&Article{},
		&PublishedArticle{},
		&Interactive{},
		&UserLikeBiz{},
		&UserCollectionBiz{},
		&Job{},
	)
}

func InitCollection(mdb *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	col := mdb.Collection("articles")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{"id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{"author_id", 1}},
		},
	})
	if err != nil {
		return err
	}
	liveCol := mdb.Collection("published_articles")
	_, err = liveCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{"id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{"author_id", 1}},
		},
	})
	return err

}

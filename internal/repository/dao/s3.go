package dao

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

type ArticleS3DAO struct {
	ArticleGORMDAO
	oss s3.S3
}

func NewArticleS3DAO(db *gorm.DB, oss s3.S3) *ArticleS3DAO {
	return &ArticleS3DAO{
		ArticleGORMDAO: ArticleGORMDAO{
			db: db,
		},
		oss: oss,
	}
}

func (a *ArticleS3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dao := NewArticleGORMDAO(tx)
		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}

		if err != nil {
			return err
		}
		art.Id = id
		now := time.Now().UnixMilli()
		pubArt := PublishedArticleV2{
			Id:       art.Id,
			Title:    art.Title,
			AuthorId: art.AuthorId,
			Ctime:    now,
			Utime:    now,
			Status:   art.Status,
		}
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  pubArt.Title,
				"utime":  now,
				"status": pubArt.Status,
			}),
		}).Create(&pubArt).Error

		return err
	})
	if err != nil {
		return 0, err
	}
	// 同步到 oss
	_, err = a.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("webook-1314583317"),
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return id, err
}

func (a *ArticleS3DAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	// 在 dao 层面同步状态（发表）
	now := time.Now().UnixMilli()
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? and author_id = ?", id, uid).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			})
		if res.Error != nil {
			return res.Error
		}
		// 防止恶意修改别人的文章
		if res.RowsAffected != 1 {
			return errors.New("更新失败，ID不对或尝试修改别人的文章")
		}
		return tx.Model(&PublishedArticleV2{}).Where("id = ?", id).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			}).Error
	})
	if err != nil {
		return err
	}
	// 处理设置未不可见的情况，此时需要将内容从 oss 中删除，还要小心 CDN 中的缓存
	const statusPrivate = 3
	if status == statusPrivate {
		_, err = a.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string]("webook-1314583317"),
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})
	}
	return err
}

// PublishedArticleV2 没有 content 字段，因为 content 存在 oss 上，不存这里
type PublishedArticleV2 struct {
	Id    int64  `gorm:"primaryKey, autoIncrement" bson:"id, omitempty"`
	Title string `gorm:"type=varchar(4096)" bson:"title, omitempty"`

	// 这里肯定是外键，但我们只给他一个索引即可，没有用数据库层面上的关联关系
	AuthorId int64 `gorm:"index" bson:"author_id, omitempty"`

	// 创建时间
	Ctime int64 `bson:"ctime, omitempty"`
	// 更新时间
	Utime int64 `bson:"utime, omitempty"`

	Status uint8 `bson:"status, omitempty"`
}

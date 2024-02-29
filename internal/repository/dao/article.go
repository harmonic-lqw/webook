package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error)
}

type ArticleGORMDAO struct {
	db *gorm.DB
}

func (a *ArticleGORMDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var res []PublishedArticle
	const ArticleStatusPublished = 2
	err := a.db.WithContext(ctx).Where("utime < ? AND status = ?", start.UnixMilli(), ArticleStatusPublished).
		Offset(offset).Limit(limit).
		First(&res).Error
	return res, err
}

func (a *ArticleGORMDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var res PublishedArticle
	err := a.db.WithContext(ctx).Where("id = ?", id).
		First(&res).Error
	return res, err

}

func (a *ArticleGORMDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := a.db.WithContext(ctx).Where("id = ?", id).
		First(&art).Error
	return art, err
}

func (a *ArticleGORMDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := a.db.WithContext(ctx).
		Where("author_id = ?", uid).
		Offset(offset).
		Limit(limit).
		// a ASC, B DESC
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (a *ArticleGORMDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	// 在 dao 层面同步状态（发表）
	now := time.Now().UnixMilli()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		return tx.Model(&PublishedArticle{}).Where("id = ?", id).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			}).Error
	})

}

func NewArticleGROMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{
		db: db,
	}
}

func (a *ArticleGORMDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dao := NewArticleGROMDAO(tx)
		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}

		if err != nil {
			return err
		}
		art.Id = id
		//err = dao.UpsertV2(ctx, PublishedArticle(art))
		now := time.Now().UnixMilli()
		pubArt := PublishedArticle(art)
		pubArt.Ctime = now
		pubArt.Utime = now
		// 直接在 sql 层面实现 upsert
		err = tx.Clauses(clause.OnConflict{
			// 对 MySQL 不起效，但是可以兼容别的方言
			// INSERT xxx ON DUPLICATE KEY SET `title`=?
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   pubArt.Title,
				"content": pubArt.Content,
				"utime":   now,
				"status":  pubArt.Status,
			}),
		}).Create(&pubArt).Error

		return err
	})

	return id, err
}

// Sync2 一种非闭包的写法
func (a *ArticleGORMDAO) Sync2(ctx context.Context, art Article) (int64, error) {
	// 在 dao 层面同步数据（发表）
	// 肯定是同库不同表，因此需要开启事务
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()

	var (
		id  = art.Id
		err error
	)

	// 复用已有代码，不然需要创建一个基于事务的 DAO 进行操作
	dao := NewArticleGROMDAO(tx)
	if id > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		id, err = dao.Insert(ctx, art)
	}

	if err != nil {
		return 0, err
	}
	art.Id = id
	//err = dao.UpsertV2(ctx, PublishedArticle(art))
	now := time.Now().UnixMilli()
	// 此处的 PublishedArticle 表示的是文章发表到线上库
	pubArt := PublishedArticle(art)
	pubArt.Ctime = now
	pubArt.Utime = now
	// 直接在 sql 层面实现 upsert
	err = tx.Clauses(clause.OnConflict{
		// 对 MySQL 不起效，但是可以兼容别的方言
		// INSERT xxx ON DUPLICATE KEY SET `title`=?
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   pubArt.Title,
			"content": pubArt.Content,
			"utime":   now,
			"status":  pubArt.Status,
		}),
	}).Create(&pubArt).Error
	if err != nil {
		return 0, err
	}

	tx.Commit()
	return id, nil
}

func (a *ArticleGORMDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&Article{}).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("更新失败，ID不对或尝试修改别人的文章")
	}
	return nil
}

func (a *ArticleGORMDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

type Article struct {
	Id      int64  `gorm:"primaryKey, autoIncrement" bson:"id, omitempty"`
	Title   string `gorm:"type=varchar(4096)" bson:"title, omitempty"`
	Content string `gorm:"type=BLOB" bson:"content, omitempty"`

	// 这里肯定是外键，但我们只给他一个索引即可，没有用数据库层面上的关联关系
	AuthorId int64 `gorm:"index" bson:"author_id, omitempty"`

	// 创建时间
	Ctime int64 `bson:"ctime, omitempty"`
	// 更新时间
	Utime int64 `bson:"utime, omitempty"`

	Status uint8 `bson:"status, omitempty"`
}

// PublishedArticle 考虑到制作库和线上库的字段不同
type PublishedArticle Article

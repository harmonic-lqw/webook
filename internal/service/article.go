package service

import (
	"context"
	"errors"
	"time"
	"webook/internal/domain"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/pkg/logger"
)

// mockgen -source .\internal\service\article.go -destination .\internal\service\mocks\article_mock.go -package svcmocks
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer

	// V1 写法，在 service 层面同步数据（发表）
	readerRepo repository.ArticleReaderRepository
	authorRepo repository.ArticleAuthorRepository

	l logger.LoggerV1
}

func (a *articleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, start, offset, limit)
}

func (a *articleService) GetPubById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)
	go func() {
		if err == nil {
			// 向 kafka 发送一个消息
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(er))
			}
		}
	}()

	return res, err
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (a *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

// NewArticleServiceV1 这里的返回值，就可以看出，返回接口和返回具体类型的在使用上的区别：返回接口无法使用 PublishV1，但返回具体类型却可以
func NewArticleServiceV1(authorRepo repository.ArticleAuthorRepository, readerRepo repository.ArticleReaderRepository, l logger.LoggerV1) *articleService {
	return &articleService{
		readerRepo: readerRepo,
		authorRepo: authorRepo,
		l:          l,
	}
}

func NewArticleService(repo repository.ArticleRepository, producer article.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	// 思路： 先操作制作库；再更新线上库
	var (
		id  = art.Id
		err error
	)
	// 有可能制作库已经有数据
	if id > 0 {
		err = a.authorRepo.Update(ctx, art)
	} else {
		id, err = a.authorRepo.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	for i := 0; i < 3; i++ {
		// 有可能线上库已经有数据了，所以 Create 不如 Save 合适
		// err = a.readerRepo.Create(ctx, art)
		err = a.readerRepo.Save(ctx, art)
		if err != nil {
			// 多接入一些 tracing 工具
			a.l.Error("保存制作库成功但是到线上库失败",
				logger.Int64("times", int64(i)),
				logger.Int64("artId", art.Id),
				logger.Error(err))
		} else {
			return id, nil
		}
	}
	a.l.Error("保存制作库成功但是到线上库失败，重试耗尽",
		logger.Int64("artId", art.Id),
		logger.Error(err))
	return id, errors.New("保存制作库成功但是到线上库失败")
}

// Save 关键点，为什么要返回文章 Id，就是在这个场景中体现：文章是新建还是更新就通过有没有 Id 来判断，因此要把这个 Id 返回给前端，下次前端发请求就带着这个 Id
func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	// 保存的时候
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	} else {
		return a.repo.Create(ctx, art)
	}
}

package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"webook/search/domain"
	"webook/search/repository/dao"
)

type articleRepository struct {
	dao   dao.ArticleDAO
	tags  dao.TagDAO
	cols  dao.CollectDAO
	likes dao.LikeESDAO
}

func (a *articleRepository) SearchArticle(ctx context.Context,
	uid int64,
	keywords []string) ([]domain.Article, error) {
	artIDs, err := a.tags.Search(ctx, uid, "article", keywords)
	if err != nil {
		return nil, err
	}

	// assignment week19
	collectIDs, err := a.cols.Search(ctx, uid, "article")
	if err != nil {
		return nil, err
	}
	likeIDs, err := a.likes.Search(ctx, uid, "article")
	if err != nil {
		return nil, err
	}

	arts, err := a.dao.Search(ctx, artIDs, collectIDs, likeIDs, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(arts, func(idx int, src dao.Article) domain.Article {
		return domain.Article{
			Id:      src.Id,
			Title:   src.Title,
			Status:  src.Status,
			Content: src.Content,
			Tags:    src.Tags,
		}
	}), nil
}

func (a *articleRepository) InputArticle(ctx context.Context, msg domain.Article) error {
	return a.dao.InputArticle(ctx, dao.Article{
		Id:      msg.Id,
		Title:   msg.Title,
		Status:  msg.Status,
		Content: msg.Content,
	})
}

func NewArticleRepository(d dao.ArticleDAO, td dao.TagDAO) ArticleRepository {
	return &articleRepository{
		dao:  d,
		tags: td,
	}
}

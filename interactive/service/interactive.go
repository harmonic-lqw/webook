package service

import (
	"context"
	"golang.org/x/sync/errgroup"
	"math"
	"time"
	"webook/interactive/domain"
	"webook/interactive/repository"
)

// mockgen -source .\internal\service\interactive.go -destination .\internal\service\mocks\interactive_mock.go -package svcmocks
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interactive, error)

	// GetTopNLike assignment week9
	GetTopNLike(ctx context.Context) ([]domain.Interactive, error)
	// SetTopNLike assignment week9
	SetTopNLike(ctx context.Context) error
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func (i *interactiveService) GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interactive, error) {
	intrs, err := i.repo.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}

func (i *interactiveService) Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error) {
	// 先拿阅读/点赞/收藏数
	intr, err := i.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}

	// 再去核对该用户对该文章是否点赞/收藏
	var eg errgroup.Group
	eg.Go(func() error {
		var er error
		intr.Liked, er = i.repo.Liked(ctx, biz, bizId, uid)
		return er
	})

	eg.Go(func() error {
		var er error
		intr.Collected, er = i.repo.Collected(ctx, biz, bizId, uid)
		return er
	})
	return intr, eg.Wait()
}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId, cid, uid int64) error {
	return i.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (i *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) LikeV1(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := i.repo.IncrLike(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	if biz != "article" {
		return nil
	}
	// 去 Redis 的 ZSET 中拿 topN 的文章
	arts, err := i.repo.GetTopNLikeZSet(ctx)
	if err != nil {
		// 记录一下日志，更新热度失败
		return nil
	}
	var index int
	var art domain.Article
	for index, art = range arts {
		if art.Id == bizId {
			break
		}
		if index == len(arts)-1 {
			// 什么都不用做，因为点赞的这篇文章不在 ZSET 中
			return nil
		}
	}

	// 更新热度
	intr, err := i.repo.Get(ctx, biz, bizId)
	if err != nil {
		// 记录一下日志，更新热度失败
		return nil
	}
	sc := i.score(intr.LikeCnt, art.Utime)
	err = i.repo.ReplaceTopNLikeZSet(ctx, sc, art)
	if err != nil {
		// 记录一下日志，更新热度失败
		return nil
	}
	return nil
}

func (i *interactiveService) score(likeCnt int64, utime time.Time) float64 {
	duration := time.Since(utime).Seconds()
	return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{repo: repo}
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

// GetTopNLike assignment week9
func (i *interactiveService) GetTopNLike(ctx context.Context) ([]domain.Interactive, error) {
	return i.repo.GetTopNLike(ctx)
}

// SetTopNLike assignment week9
func (i *interactiveService) SetTopNLike(ctx context.Context) error {
	return i.repo.SetTopNLike(ctx)
}

package service

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
)

type RankingService interface {
	// TopN 前100
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	// 用于获取点赞数
	intrSvc InteractiveService
	// 用于查找文章
	artSvc ArticleService

	batchSize int
	scoreFunc func(likeCnt int64, utime time.Time) float64
	// top 几
	n int

	repo repository.RankingRepository
}

func NewBatchRankingService(intrSvc InteractiveService, artSvc ArticleService, repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		intrSvc:   intrSvc,
		artSvc:    artSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
		repo: repo,
	}
}

func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}
func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	// 如果数据库出了问题，定时任务失败，不可能更新到缓存
	if err != nil {
		return err
	}
	// 存到缓存中
	return b.repo.ReplaceTopN(ctx, arts)
}

// topN 算法本身，通过它可以更好的进行测试
func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	// 一个可以优化的点：太久的文章不太可能会是热榜，因此只查询 7 天内的
	ddl := start.Add(-7 * 24 * time.Hour)

	// 一个优先级队列
	type Score struct {
		score float64
		art   domain.Article
	}

	topN := queue.NewPriorityQueue[Score](b.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score == dst.score {
			return 0
		} else {
			return -1
		}

	})

	for {
		// 取文章数据
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		if len(arts) <= 0 {
			break
		}
		ids := slice.Map(arts, func(idx int, art domain.Article) int64 {
			return art.Id
		})
		// 取点赞数
		intrMap, err := b.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		for _, art := range arts {
			intr := intrMap[art.Id]
			score := b.scoreFunc(intr.LikeCnt, art.Utime)
			ele := Score{
				score: score,
				art:   art,
			}
			err = topN.Enqueue(ele)
			if err == queue.ErrOutOfCapacity {
				// 满了，需要替换
				minEle, _ := topN.Dequeue()
				if minEle.score < score {
					_ = topN.Enqueue(ele)
				} else {
					_ = topN.Enqueue(minEle)
				}
			}
		}
		offset += len(arts)
		// 如果某一批次没有取满，说明没有下一批
		if len(arts) < b.batchSize ||
			arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}

	// 此时 topN 里就是最终数据
	res := make([]domain.Article, topN.Len())
	// 从后往前装
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}

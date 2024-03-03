package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	domain2 "webook/interactive/domain"
	"webook/interactive/service"
	"webook/internal/domain"
	svcmocks "webook/internal/service/mocks"
)

func TestBatchRankingService_TopN(t *testing.T) {
	const batchSize = 2
	now := time.Now()
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.InteractiveService, ArticleService)

		wantArts []domain.Article
		wantErr  error
	}{
		{
			name: "成功获取",
			mock: func(ctrl *gomock.Controller) (service.InteractiveService, ArticleService) {
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				artSvc := svcmocks.NewMockArticleService(ctrl)
				// 模拟数据库的分批查询
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 2).
					Return([]domain.Article{
						{Id: 1, Utime: now},
						{Id: 2, Utime: now},
					}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, 2).
					Return([]domain.Article{
						{Id: 3, Utime: now},
						{Id: 4, Utime: now},
					}, nil)
				// 查询到第三批没数据了
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 4, 2).
					Return([]domain.Article{}, nil)

				// 第一批点赞数据
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2}).
					Return(map[int64]domain2.Interactive{
						1: {LikeCnt: 12},
						2: {LikeCnt: 23},
					}, nil)
				// 第二批点赞数据
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{3, 4}).
					Return(map[int64]domain2.Interactive{
						3: {LikeCnt: 34},
						4: {LikeCnt: 45},
					}, nil)
				return intrSvc, artSvc
			},
			wantErr: nil,
			// 预期取出 Top3 ，且从大到小排序
			wantArts: []domain.Article{
				{Id: 4, Utime: now},
				{Id: 3, Utime: now},
				{Id: 2, Utime: now},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			intrSvc, artSvc := tc.mock(ctrl)
			svc := &BatchRankingService{
				intrSvc:   intrSvc,
				artSvc:    artSvc,
				batchSize: batchSize,
				scoreFunc: func(likeCnt int64, utime time.Time) float64 {
					// 一种简化的计算热榜的方式，不影响测试
					return float64(likeCnt)
				},
				n: 3,
			}
			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantArts, arts)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

package async

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
	"webook/internal/service/sms"
	smsmocks "webook/internal/service/sms/mocks"
	"webook/pkg/logger"
	loggermocks "webook/pkg/logger/mocks"
)

func TestAsyncService_Send(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (sms.Service, repository.AsyncSmsRepository, logger.LoggerV1)

		signAsync bool
		errCnt    int32
		reqCnt    int32
		keep      int32

		wantSignAsync bool
		wantErrCnt    int32
		wantReqCnt    int32
		wantKeep      int32

		wantErr error
	}{
		{
			name: "异步发送成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.AsyncSmsRepository, logger.LoggerV1) {
				svc := smsmocks.NewMockService(ctrl)
				repo := repomocks.NewMockAsyncSmsRepository(ctrl)
				logV1 := loggermocks.NewMockLoggerV1(ctrl)
				repo.EXPECT().Add(gomock.Any(), domain.AsyncSms{
					TplId:   "123",
					Args:    []string{"234"},
					Numbers: []string{"345"},
					// 重试的最大次数
					RetryMax: 3,
				}).
					Return(nil)
				repo.EXPECT().PreemptWaitingSMS(gomock.Any()).
					Return(domain.AsyncSms{
						TplId:   "123",
						Args:    []string{"234"},
						Numbers: []string{"345"},
						// 重试的最大次数
						RetryMax: 3,
					}, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().ReportScheduleResult(gomock.Any(), gomock.Any(), true).Return(nil)

				return svc, repo, logV1
			},

			signAsync: true,
			errCnt:    0,
			reqCnt:    0,
			keep:      10,

			wantSignAsync: true,
			wantErrCnt:    0,
			wantReqCnt:    0,
			wantKeep:      10,
			wantErr:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc, repo, logV1 := tc.mock(ctrl)

			asvc := NewService(svc, repo, logV1)
			asvc.signAsync = tc.signAsync
			asvc.errCnt = tc.errCnt
			asvc.reqCnt = tc.reqCnt
			asvc.keep = tc.keep

			err := asvc.Send(context.Background(), "123", []string{"234"}, "345")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantSignAsync, asvc.signAsync)
			assert.Equal(t, tc.wantReqCnt, asvc.reqCnt)
			assert.Equal(t, tc.wantErrCnt, asvc.errCnt)
			assert.Equal(t, tc.wantKeep, asvc.keep)
		})
	}
}

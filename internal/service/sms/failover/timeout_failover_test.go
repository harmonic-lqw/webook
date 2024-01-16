package failover

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/service/sms"
	smsmocks "webook/internal/service/sms/mocks"
)

func TestTimeoutFailoverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name      string
		mocks     func(ctrl *gomock.Controller) []sms.Service
		threshold int32

		idx int32
		cnt int32

		wantErr error
		wantCnt int32
		wantIdx int32
	}{
		{
			name: "没有触发切换",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0}
			},

			idx:       0,
			cnt:       12,
			threshold: 15,

			wantErr: nil,
			wantIdx: 0,
			wantCnt: 0,
		},
		{
			name: "触发切换，成功",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},

			idx:       0,
			cnt:       16,
			threshold: 15,

			wantErr: nil,
			wantIdx: 1,
			wantCnt: 0,
		},
		{
			// 触发切换后，新的服务商无法发送，但不是超时错误
			name: "触发切换，失败",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("切换后的服务商发送失败"))
				return []sms.Service{svc0, svc1}
			},

			idx:       1,
			cnt:       16,
			threshold: 15,

			wantErr: errors.New("切换后的服务商发送失败"),
			wantIdx: 0,
			wantCnt: 0,
		},
		{
			name: "触发切换，超时",
			mocks: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smsmocks.NewMockService(ctrl)
				svc1 := smsmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				return []sms.Service{svc0, svc1}
			},

			idx:       1,
			cnt:       16,
			threshold: 15,

			wantErr: context.DeadlineExceeded,
			wantIdx: 0,
			wantCnt: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svcs := tc.mocks(ctrl)
			svc := NewTimeoutFailoverSMSService(svcs, tc.threshold)
			svc.cnt = tc.cnt
			svc.idx = tc.idx
			err := svc.Send(context.Background(), "123", []string{"234"}, "345")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantCnt, svc.cnt)
			assert.Equal(t, tc.wantIdx, svc.idx)
		})
	}
}

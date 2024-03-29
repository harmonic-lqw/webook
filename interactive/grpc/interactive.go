package grpc

import (
	"context"
	grpc2 "google.golang.org/grpc"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/interactive/domain"
	"webook/interactive/service"
)

type InteractiveServiceServer struct {
	intrv1.UnimplementedInteractiveServiceServer
	svc service.InteractiveService
}

func (i *InteractiveServiceServer) Register(s *grpc2.Server) {
	intrv1.RegisterInteractiveServiceServer(s, i)
}

func NewInteractiveServiceServer(svc service.InteractiveService) *InteractiveServiceServer {
	return &InteractiveServiceServer{svc: svc}
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intrv1.IncrReadCntRequest) (*intrv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	if err != nil {
		return nil, err
	}
	return &intrv1.IncrReadCntResponse{}, nil
}

func (i *InteractiveServiceServer) Like(ctx context.Context, request *intrv1.LikeRequest) (*intrv1.LikeResponse, error) {
	err := i.svc.Like(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv1.LikeResponse{}, nil
}

func (i *InteractiveServiceServer) CancelLike(ctx context.Context, request *intrv1.CancelLikeRequest) (*intrv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv1.CancelLikeResponse{}, nil
}

func (i *InteractiveServiceServer) Collect(ctx context.Context, request *intrv1.CollectRequest) (*intrv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.GetBiz(), request.GetBizId(), request.GetCid(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv1.CollectResponse{}, nil

}

func (i *InteractiveServiceServer) Get(ctx context.Context, request *intrv1.GetRequest) (*intrv1.GetResponse, error) {
	intr, err := i.svc.Get(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv1.GetResponse{
		Intr: i.toDTO(intr),
	}, nil
}

func (i *InteractiveServiceServer) GetByIds(ctx context.Context, request *intrv1.GetByIdsRequest) (*intrv1.GetByIdsResponse, error) {
	res, err := i.svc.GetByIds(ctx, request.GetBiz(), request.GetIds())
	if err != nil {
		return nil, err
	}

	intrs := make(map[int64]*intrv1.Interactive, len(res))
	for key, val := range res {
		intrs[key] = i.toDTO(val)
	}
	return &intrv1.GetByIdsResponse{
		Intrs: intrs,
	}, nil
}

func (i *InteractiveServiceServer) mustEmbedUnimplementedInteractiveServiceServer() {
	//TODO implement me
	panic("implement me")
}

// DTO：数据传输对象，这里指的是 protobuf 定义的结构体
func (i *InteractiveServiceServer) toDTO(intr domain.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		CollectCnt: intr.CollectCnt,
		LikeCnt:    intr.LikeCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
}

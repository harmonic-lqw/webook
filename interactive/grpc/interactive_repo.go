package grpc

import (
	"context"
	"google.golang.org/grpc"
	intrv2 "webook/api/proto/gen/intr/v2"
	"webook/interactive/domain"
	"webook/interactive/repository"
)

type InteractiveRepositoryServer struct {
	intrv2.UnimplementedInteractiveRepositoryServer
	repo repository.InteractiveRepository
}

func NewInteractiveRepositoryServer(repo repository.InteractiveRepository) *InteractiveRepositoryServer {
	return &InteractiveRepositoryServer{repo: repo}
}

func (i *InteractiveRepositoryServer) IncrReadCnt(ctx context.Context, request *intrv2.IncrReadCntRequest) (*intrv2.IncrReadCntResponse, error) {
	err := i.repo.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	if err != nil {
		return nil, err
	}
	return &intrv2.IncrReadCntResponse{}, nil
}

func (i *InteractiveRepositoryServer) BatchIncrReadCnt(ctx context.Context, request *intrv2.BatchIncrReadCntRequest) (*intrv2.BatchIncrReadCntResponse, error) {
	err := i.repo.BatchIncrReadCnt(ctx, request.GetBizs(), request.GetBizIds())
	if err != nil {
		return nil, err
	}
	return &intrv2.BatchIncrReadCntResponse{}, nil
}

func (i *InteractiveRepositoryServer) IncrLike(ctx context.Context, request *intrv2.IncrLikeRequest) (*intrv2.IncrLikeResponse, error) {
	err := i.repo.IncrLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.IncrLikeResponse{}, nil
}

func (i *InteractiveRepositoryServer) DecrLike(ctx context.Context, request *intrv2.DecrLikeRequest) (*intrv2.DecrLikeResponse, error) {
	err := i.repo.DecrLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.DecrLikeResponse{}, nil
}

func (i *InteractiveRepositoryServer) AddCollectionItem(ctx context.Context, request *intrv2.AddCollectionItemRequest) (*intrv2.AddCollectionItemResponse, error) {
	err := i.repo.AddCollectionItem(ctx, request.GetBiz(), request.GetBizId(), request.GetCid(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.AddCollectionItemResponse{}, nil
}

func (i *InteractiveRepositoryServer) Get(ctx context.Context, request *intrv2.GetRequest) (*intrv2.GetResponse, error) {
	resp, err := i.repo.Get(ctx, request.GetBiz(), request.GetBizId())
	if err != nil {
		return nil, err
	}
	return &intrv2.GetResponse{
		Intr: i.toDTO(resp),
	}, nil
}

func (i *InteractiveRepositoryServer) Liked(ctx context.Context, request *intrv2.LikedRequest) (*intrv2.LikedResponse, error) {
	resp, err := i.repo.Liked(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.LikedResponse{
		Liked: resp,
	}, nil
}

func (i *InteractiveRepositoryServer) Collected(ctx context.Context, request *intrv2.CollectedRequest) (*intrv2.CollectedResponse, error) {
	resp, err := i.repo.Collected(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.CollectedResponse{
		Collected: resp,
	}, nil
}

func (i *InteractiveRepositoryServer) GetByIds(ctx context.Context, request *intrv2.GetByIdsRequest) (*intrv2.GetByIdsResponse, error) {
	resp, err := i.repo.GetByIds(ctx, request.GetBiz(), request.GetBizIds())
	if err != nil {
		return nil, err
	}
	intrs := make([]*intrv2.Interactive, len(resp))
	for j := 0; j < len(resp); j++ {
		intrs = append(intrs, i.toDTO(resp[j]))
	}
	return &intrv2.GetByIdsResponse{
		Intrs: intrs,
	}, nil
}

func (i *InteractiveRepositoryServer) toDTO(intr domain.Interactive) *intrv2.Interactive {
	return &intrv2.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
}

func (i *InteractiveRepositoryServer) Register(server *grpc.Server) {
	intrv2.RegisterInteractiveRepositoryServer(server, i)
}

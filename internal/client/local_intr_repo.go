package client

import (
	"context"
	"google.golang.org/grpc"
	intrv2 "webook/api/proto/gen/intr/v2"
	"webook/interactive/domain"
	"webook/interactive/repository"
)

type LocalInteractiveRepositoryClient struct {
	repo repository.InteractiveRepository
}

func NewLocalInteractiveRepositoryClient(repo repository.InteractiveRepository) *LocalInteractiveRepositoryClient {
	return &LocalInteractiveRepositoryClient{repo: repo}
}

func (l *LocalInteractiveRepositoryClient) IncrReadCnt(ctx context.Context, in *intrv2.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv2.IncrReadCntResponse, error) {
	err := l.repo.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	if err != nil {
		return nil, err
	}
	return &intrv2.IncrReadCntResponse{}, nil
}

func (l *LocalInteractiveRepositoryClient) BatchIncrReadCnt(ctx context.Context, in *intrv2.BatchIncrReadCntRequest, opts ...grpc.CallOption) (*intrv2.BatchIncrReadCntResponse, error) {
	err := l.repo.BatchIncrReadCnt(ctx, in.GetBizs(), in.GetBizIds())
	if err != nil {
		return nil, err
	}
	return &intrv2.BatchIncrReadCntResponse{}, nil
}

func (l *LocalInteractiveRepositoryClient) IncrLike(ctx context.Context, in *intrv2.IncrLikeRequest, opts ...grpc.CallOption) (*intrv2.IncrLikeResponse, error) {
	err := l.repo.IncrLike(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.IncrLikeResponse{}, nil
}

func (l *LocalInteractiveRepositoryClient) DecrLike(ctx context.Context, in *intrv2.DecrLikeRequest, opts ...grpc.CallOption) (*intrv2.DecrLikeResponse, error) {
	err := l.repo.DecrLike(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.DecrLikeResponse{}, nil
}

func (l *LocalInteractiveRepositoryClient) AddCollectionItem(ctx context.Context, in *intrv2.AddCollectionItemRequest, opts ...grpc.CallOption) (*intrv2.AddCollectionItemResponse, error) {
	err := l.repo.AddCollectionItem(ctx, in.GetBiz(), in.GetBizId(), in.GetCid(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.AddCollectionItemResponse{}, nil
}

func (l *LocalInteractiveRepositoryClient) Get(ctx context.Context, in *intrv2.GetRequest, opts ...grpc.CallOption) (*intrv2.GetResponse, error) {
	resp, err := l.repo.Get(ctx, in.GetBiz(), in.GetBizId())
	if err != nil {
		return nil, err
	}

	return &intrv2.GetResponse{
		Intr: l.toDTO(resp),
	}, nil
}

func (l *LocalInteractiveRepositoryClient) Liked(ctx context.Context, in *intrv2.LikedRequest, opts ...grpc.CallOption) (*intrv2.LikedResponse, error) {
	resp, err := l.repo.Liked(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.LikedResponse{
		Liked: resp,
	}, nil
}

func (l *LocalInteractiveRepositoryClient) Collected(ctx context.Context, in *intrv2.CollectedRequest, opts ...grpc.CallOption) (*intrv2.CollectedResponse, error) {
	resp, err := l.repo.Collected(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv2.CollectedResponse{
		Collected: resp,
	}, nil
}

func (l *LocalInteractiveRepositoryClient) GetByIds(ctx context.Context, in *intrv2.GetByIdsRequest, opts ...grpc.CallOption) (*intrv2.GetByIdsResponse, error) {
	resp, err := l.repo.GetByIds(ctx, in.GetBiz(), in.GetBizIds())
	if err != nil {
		return nil, err
	}
	intrs := make([]*intrv2.Interactive, len(resp))
	for i := 0; i < len(resp); i++ {
		intrs = append(intrs, l.toDTO(resp[i]))
	}
	return &intrv2.GetByIdsResponse{
		Intrs: intrs,
	}, nil
}

func (l *LocalInteractiveRepositoryClient) toDTO(intr domain.Interactive) *intrv2.Interactive {
	return &intrv2.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		CollectCnt: intr.CollectCnt,
		LikeCnt:    intr.LikeCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
}

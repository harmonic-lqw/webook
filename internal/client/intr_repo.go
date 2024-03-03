package client

import (
	"context"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"google.golang.org/grpc"
	"math/rand"
	intrv2 "webook/api/proto/gen/intr/v2"
)

type InteractiveRepositoryClient struct {
	remote intrv2.InteractiveRepositoryClient
	local  *LocalInteractiveRepositoryClient

	threshold *atomicx.Value[int32]
}

func NewInteractiveRepositoryClient(remote intrv2.InteractiveRepositoryClient, local *LocalInteractiveRepositoryClient) *InteractiveRepositoryClient {
	return &InteractiveRepositoryClient{
		remote:    remote,
		local:     local,
		threshold: atomicx.NewValue[int32](),
	}
}

func (i *InteractiveRepositoryClient) IncrReadCnt(ctx context.Context, in *intrv2.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv2.IncrReadCntResponse, error) {
	return i.selectClient().IncrReadCnt(ctx, in, opts...)
}

func (i *InteractiveRepositoryClient) BatchIncrReadCnt(ctx context.Context, in *intrv2.BatchIncrReadCntRequest, opts ...grpc.CallOption) (*intrv2.BatchIncrReadCntResponse, error) {
	return i.selectClient().BatchIncrReadCnt(ctx, in, opts...)
}

func (i *InteractiveRepositoryClient) IncrLike(ctx context.Context, in *intrv2.IncrLikeRequest, opts ...grpc.CallOption) (*intrv2.IncrLikeResponse, error) {
	return i.selectClient().IncrLike(ctx, in, opts...)
}

func (i *InteractiveRepositoryClient) DecrLike(ctx context.Context, in *intrv2.DecrLikeRequest, opts ...grpc.CallOption) (*intrv2.DecrLikeResponse, error) {
	return i.selectClient().DecrLike(ctx, in, opts...)
}

func (i *InteractiveRepositoryClient) AddCollectionItem(ctx context.Context, in *intrv2.AddCollectionItemRequest, opts ...grpc.CallOption) (*intrv2.AddCollectionItemResponse, error) {
	return i.selectClient().AddCollectionItem(ctx, in, opts...)
}

func (i *InteractiveRepositoryClient) Get(ctx context.Context, in *intrv2.GetRequest, opts ...grpc.CallOption) (*intrv2.GetResponse, error) {
	return i.selectClient().Get(ctx, in, opts...)
}

func (i *InteractiveRepositoryClient) Liked(ctx context.Context, in *intrv2.LikedRequest, opts ...grpc.CallOption) (*intrv2.LikedResponse, error) {
	return i.selectClient().Liked(ctx, in, opts...)
}

func (i *InteractiveRepositoryClient) Collected(ctx context.Context, in *intrv2.CollectedRequest, opts ...grpc.CallOption) (*intrv2.CollectedResponse, error) {
	return i.selectClient().Collected(ctx, in, opts...)
}

func (i *InteractiveRepositoryClient) GetByIds(ctx context.Context, in *intrv2.GetByIdsRequest, opts ...grpc.CallOption) (*intrv2.GetByIdsResponse, error) {
	return i.selectClient().GetByIds(ctx, in, opts...)
}

func (i *InteractiveRepositoryClient) selectClient() intrv2.InteractiveRepositoryClient {
	// 随机 1-100
	num := rand.Int31n(100)
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

func (i *InteractiveRepositoryClient) UpdateThreshold(val int32) {
	i.threshold.Store(val)
}

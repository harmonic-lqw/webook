package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	// 创建一个 gRPC Server
	gs := grpc.NewServer()
	// 创建一个 UserServiceServer 并注册
	us := &Server{}
	RegisterUserServiceServer(gs, us)

	// 创建一个监听网络端口的 listener
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	// 调用 gRPC Server 上的 Serve 方法，开始监听
	err = gs.Serve(l)
	t.Log(err)
}

func TestClient(t *testing.T) {
	// 初始化一个连接池
	cc, err := grpc.Dial("localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	// 用该连接池初始化一个客户端
	client := NewUserServiceClient(cc)
	// 发起调用
	resp, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 12})
	require.NoError(t, err)
	t.Log(resp.User)
}

package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
	"webook/pkg/grpcx/interceptor/trace"
)

type InterceptorTestSuite struct {
	suite.Suite
}

func (s *InterceptorTestSuite) TestServer() {
	t := s.T()
	InitOTEL()
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(NewLogInterceptor(t),
			trace.NewOTELInterceptorBuilder("server_test", nil, nil).
				BuildUnaryServerInterceptor()))
	RegisterUserServiceServer(server, &Server{
		Name: "interceptor_test",
	})

	// 使用装饰器进行针对业务进行限流
	//RegisterUserServiceServer(server, &LimiterUserServer{
	//	UserServiceServer: &Server{
	//		Name: "interceptor_test",
	//	},
	//})

	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	_ = server.Serve(l)
}

func (s *InterceptorTestSuite) TestClient() {
	t := s.T()
	InitOTEL()
	// 初始化一个连接池
	cc, err := grpc.Dial("localhost:8090",
		grpc.WithUnaryInterceptor(trace.NewOTELInterceptorBuilder("client_test", nil, nil).BuildUnaryClientInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	// 用该连接池初始化一个客户端
	client := NewUserServiceClient(cc)
	// 发起调用
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 设置 grpcHeadValue，客户端便可以通过 grpcHeadValue 拿到
	//md, ok := metadata.FromIncomingContext(ctx)
	//if !ok {
	//	md = metadata.New(make(map[string]string))
	//}
	//md.Set("app", "test_client")

	resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 12})
	require.NoError(t, err)
	t.Log(resp.User)
	time.Sleep(time.Second)
}

func TestInterceptorTestSuite(t *testing.T) {
	suite.Run(t, new(InterceptorTestSuite))
}

func NewLogInterceptor(t *testing.T) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		t.Log("请求处理前", req, info)
		resp, err = handler(ctx, req)
		t.Log("请求处理后", resp, err)
		return
	}
}

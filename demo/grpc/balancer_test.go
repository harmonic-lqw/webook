package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/weightedroundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
	_ "webook/pkg/grpcx/balancer/wrr"
)

type BalancerTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *BalancerTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	require.NoError(s.T(), err)
	s.client = cli
}

func (s *BalancerTestSuite) TestServer() {
	go func() {
		s.startEtcdServer(":8090", 10, &Server{
			Name: ":8090",
		})
	}()
	go func() {
		s.startEtcdServer(":8091", 20, &Server{
			Name: ":8091",
		})
	}()
	s.startEtcdServer(":8092", 30, &FailedServer{
		Name: ":8092",
	})
}

func (s *BalancerTestSuite) startEtcdServer(addr string, weight int, svr UserServiceServer) {
	t := s.T()
	em, err := endpoints.NewManager(s.client, "service/user")
	require.NoError(t, err)

	// 1.先打开服务端口
	l, err := net.Listen("tcp", "127.0.0.1"+addr)
	require.NoError(s.T(), err)

	// 租约机制
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var ttl int64 = 5 // 单位是秒
	leaseResp, err := s.client.Grant(ctx, ttl)
	require.NoError(t, err)

	// 2.再服务注册（指注册到服务中心)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr = "127.0.0.1" + addr
	key := "service/user/" + addr
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息
		Addr: addr,
		Metadata: map[string]any{
			"weight": weight,
		},
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)

	// 续约
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		_, err1 := s.client.KeepAlive(kaCtx, leaseResp.ID)
		// 没有 err1 就代表续约成功
		require.NoError(t, err1)
		//for kaResp := range ch {
		//	t.Log(kaResp.String())
		//}
	}()

	// 这里的注册是开启微服务
	server := grpc.NewServer()
	RegisterUserServiceServer(server, svr)
	server.Serve(l)

	// 退出续约
	kaCancel()

	// 服务删除（不过如果服务器宕机，根本执行不到）
	err = em.DeleteEndpoint(ctx, key)
	require.NoError(t, err)
	// 服务器优雅退出（微服务框架 grpc 已经帮你实现好了）
	server.GracefulStop()

}

func (s *BalancerTestSuite) TestEtcdFailOverClient() {
	t := s.T()
	// 不需要关心数据同步问题，因为这个 resolver 已经帮你去监听了注册信息的变动了
	etcdResolver, err := resolver.NewBuilder(s.client)
	require.NoError(t, err)

	// 去 etcd 找服务端
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`
{
  "loadBalancingConfig": [{"round_robin":  {}}],
  "methodConfig":  [
    {
      "name": [{"service": "UserService"}],
      "retryPolicy": {
        "maxAttempts": 4,
        "initialBackoff": "0.01s",
        "maxBackoff": "0.1s",
        "backoffMultiplier": 2.0,
        "retryableStatusCodes": ["UNAVAILABLE"]
      }
    }
  ]
}
`))
	require.NoError(t, err)
	// 用该连接池初始化一个客户端
	client := NewUserServiceClient(cc)
	// 发起调用
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 12})
		require.NoError(t, err)
		t.Log(resp.User)
		cancel()
	}
}

func (s *BalancerTestSuite) TestEtcdClient() {
	t := s.T()
	// 不需要关心数据同步问题，因为这个 resolver 已经帮你去监听了注册信息的变动了
	etcdResolver, err := resolver.NewBuilder(s.client)
	require.NoError(t, err)

	// 去 etcd 找服务端
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc 默认只会选择第一个节点，即 peek first
		// 开启轮询算法
		//grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [ { "round_robin": {} } ]}`),
	)
	require.NoError(t, err)
	// 用该连接池初始化一个客户端
	client := NewUserServiceClient(cc)
	// 发起调用
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 12})
		require.NoError(t, err)
		t.Log(resp.User)
		cancel()
	}
}

func (s *BalancerTestSuite) TestEtcdClientWRR() {
	t := s.T()
	// 不需要关心数据同步问题，因为这个 resolver 已经帮你去监听了注册信息的变动了
	etcdResolver, err := resolver.NewBuilder(s.client)
	require.NoError(t, err)

	// 去 etcd 找服务端
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc 默认只会选择第一个节点，即 peek first
		// 开启加权轮询算法
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [ { "weighted_round_robin": {} } ]}`),
	)
	require.NoError(t, err)
	// 用该连接池初始化一个客户端
	client := NewUserServiceClient(cc)
	// 发起调用
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 12})
		require.NoError(t, err)
		t.Log(resp.User)
		cancel()
	}
}

func (s *BalancerTestSuite) TestEtcdClientCustom() {
	t := s.T()
	// 不需要关心数据同步问题，因为这个 resolver 已经帮你去监听了注册信息的变动了
	etcdResolver, err := resolver.NewBuilder(s.client)
	require.NoError(t, err)

	// 去 etcd 找服务端
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`
{
  "loadBalancingConfig": [{"custom_weighted_round_robin":  {}}],
  "methodConfig":  [
    {
      "name": [{"service": "UserService"}],
      "retryPolicy": {
        "maxAttempts": 4,
        "initialBackoff": "0.01s",
        "maxBackoff": "0.1s",
        "backoffMultiplier": 2.0,
        "retryableStatusCodes": ["UNAVAILABLE"]
      }
    }
  ]
}
`),
	)
	require.NoError(t, err)
	// 用该连接池初始化一个客户端
	client := NewUserServiceClient(cc)
	// 发起调用
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 12})
		require.NoError(t, err)
		t.Log(resp.User)
		cancel()
	}
}

func TestBalance(t *testing.T) {
	suite.Run(t, new(BalancerTestSuite))
}

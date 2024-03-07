package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	require.NoError(s.T(), err)
	s.client = cli
}

func (s *EtcdTestSuite) TestEtcdServer() {
	t := s.T()
	em, err := endpoints.NewManager(s.client, "service/user")
	require.NoError(t, err)

	// 1.先打开服务端口
	l, err := net.Listen("tcp", "127.0.0.1:8090")
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
	addr := "127.0.0.1:8090"
	key := "service/user/" + addr
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)

	// 续约
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		ch, err1 := s.client.KeepAlive(kaCtx, leaseResp.ID)
		// 没有 err1 就代表续约成功
		require.NoError(t, err1)
		for kaResp := range ch {
			t.Log(kaResp.String())
		}
	}()

	// 模拟注册信息变动
	go func() {
		for i := 0; i < 3; i++ {
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			err1 := em.Update(ctx1, []*endpoints.UpdateWithOpts{
				{
					Update: endpoints.Update{
						Op:  endpoints.Add,
						Key: key,
						Endpoint: endpoints.Endpoint{
							Addr:     addr,
							Metadata: i,
						},
					},
					Opts: []etcdv3.OpOption{etcdv3.WithLease(leaseResp.ID)},
				},
			})
			//err1 := em.AddEndpoint(ctx1, key, endpoints.Endpoint{
			//	Addr:     addr,
			//	Metadata: i,
			//}, etcdv3.WithLease(leaseResp.ID))
			cancel1()
			if err1 != nil {
				t.Log(err1)
			}
			time.Sleep(time.Second * 5)
		}
	}()

	// 这里的注册是开启微服务
	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	server.Serve(l)

	// 退出续约
	kaCancel()

	// 服务删除（不过如果服务器宕机，根本执行不到）
	err = em.DeleteEndpoint(ctx, key)
	require.NoError(t, err)
	// 服务器优雅退出（微服务框架 grpc 已经帮你实现好了）
	server.GracefulStop()
	// etcd 客户端（注册中心）关闭
	s.client.Close()

}

func (s *EtcdTestSuite) TestEtcdClient() {
	t := s.T()
	// 不需要关心数据同步问题，因为这个 resolver 已经帮你去监听了注册信息的变动了
	etcdResolver, err := resolver.NewBuilder(s.client)
	require.NoError(t, err)

	// 去 etcd 找服务端
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	// 用该连接池初始化一个客户端
	client := NewUserServiceClient(cc)
	// 发起调用
	resp, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 12})
	require.NoError(t, err)
	t.Log(resp.User)
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}

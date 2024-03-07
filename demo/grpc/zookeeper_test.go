package grpc

import (
	"context"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type ZookeeperTestSuite struct {
	suite.Suite
	conn *zk.Conn
}

func (z *ZookeeperTestSuite) SetupSuite() {
	// 连接 zookeeper
	conn, _, err := zk.Connect([]string{"127.0.0.1:2181"}, time.Second)
	require.NoError(z.T(), err)
	z.conn = conn
}

func (z *ZookeeperTestSuite) TestZookeeperServer() {
	t := z.T()

	l, err := net.Listen("tcp", "127.0.0.1:8090")
	require.NoError(t, err)

	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})

	time.Sleep(time.Second)
	// 服务注册
	// 服务节点路径，唯一标识注册的服务，类似于 etcd 中的 key
	path := "/service/user" + "/" + "127.0.0.1:8090"
	data := []byte("127.0.0.1:8090")
	_, err = z.conn.Create(path, data, zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	require.NoError(t, err)

	err = server.Serve(l)
	require.NoError(t, err)
}

func (z *ZookeeperTestSuite) TestZookeeperClient() {
	t := z.T()
	nodes, _, err := z.conn.Children("/service/user")
	require.NoError(t, err)
	for _, node := range nodes {
		nodePath := "/service/user" + "/" + node
		data, _, er := z.conn.Get(nodePath)
		require.NoError(t, er)

		cc, er := grpc.Dial(string(data),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		// 用该连接池初始化一个客户端
		client := NewUserServiceClient(cc)
		fmt.Printf("Calling service on node %s with data: %s\n", node, string(data))
		// 在这里添加调用服务的具体逻辑
		resp, er := client.GetByID(context.Background(), &GetByIDRequest{Id: 1023})
		require.NoError(t, er)
		t.Log(resp.User)
	}
}

func TestZookeeper(t *testing.T) {
	suite.Run(t, new(ZookeeperTestSuite))
}

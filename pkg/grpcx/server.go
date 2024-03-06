package grpcx

import (
	"context"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
	"webook/pkg/logger"
	"webook/pkg/netx"
)

// Server 给 grpc.Server 做了一个封装
type Server struct {
	*grpc.Server
	Port     int
	EtcdAddr string
	Name     string
	client   *etcdv3.Client
	KaCancel context.CancelFunc

	L logger.LoggerV1
}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	// 完成服务注册
	err = s.registerEtcd()
	if err != nil {
		return err
	}

	return s.Server.Serve(l)
}

func (s *Server) registerEtcd() error {
	etcdClient, err := etcdv3.NewFromURL(s.EtcdAddr)
	if err != nil {
		return err
	}
	s.client = etcdClient

	em, err := endpoints.NewManager(s.client, "service/"+s.Name)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var ttl int64 = 5 // 单位是秒
	leaseResp, err := s.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	// 续约
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.KaCancel = kaCancel
	ch, err := s.client.KeepAlive(kaCtx, leaseResp.ID)
	go func() {
		for kaResp := range ch {
			s.L.Debug(kaResp.String())
		}
	}()

	return err
}

func (s *Server) Close() error {
	if s.KaCancel != nil {
		s.KaCancel()
	}
	if s.client != nil {
		// 如果采用依赖注入的形式初始化 etcd 客户端，就不需要我去关了
		return s.client.Close()
	}
	s.GracefulStop()
	return nil
}

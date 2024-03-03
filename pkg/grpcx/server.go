package grpcx

import (
	"google.golang.org/grpc"
	"net"
)

// Server 给 grpc.Server 做了一个封装
type Server struct {
	*grpc.Server
	Addr string
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		panic(err)
	}
	return s.Server.Serve(l)
}

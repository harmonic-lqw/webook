package grpc

import (
	"context"
	"go.opentelemetry.io/otel"
	"log"
	"time"
)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

func (s *Server) GetByID(ctx context.Context, request *GetByIDRequest) (*GetByIDResponse, error) {
	ctx, span := otel.Tracer("server_biz").Start(ctx, "get_by_id")
	defer span.End()
	ddl, ok := ctx.Deadline()
	if ok {
		rest := ddl.Sub(time.Now())
		log.Println(rest.String())
	} else {
		log.Println(ddl)
	}
	time.Sleep(time.Millisecond * 100)
	return &GetByIDResponse{
		User: &User{
			Id:   123,
			Name: "name: " + s.Name,
		},
	}, nil
}

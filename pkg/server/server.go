package server

import (
	"context"
	"log"
	"net"
	"tuber/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Serve serves
func Serve() {
	lis, err := net.Listen("tcp", ":9000")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// s := chat.Server{}
	grpcServer := grpc.NewServer()

	s := Server{}

	proto.RegisterTuberServiceServer(grpcServer, &s)
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

type Server struct {
}

func (s *Server) CreateReviewApp(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	var err error

	res := proto.Response{
		Hostname: "example",
		Error:    "",
	}

	return &res, err
}

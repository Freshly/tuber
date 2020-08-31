package server

import (
	"context"
	"log"
	"net"
	"tuber/pkg/k8s"
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

// Server serves
type Server struct {
}

// Authenticate authenticates
func (s *Server) Authenticate(appName, token string) bool {
	return k8s.CanDeploy(appName, token)
}

// CreateReviewApp creates a review app
func (s *Server) CreateReviewApp(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	var err error

	authenticated := s.Authenticate(req.GetAppName(), req.GetToken())

	if authenticated {
		// Create review app

	} else {
		res := proto.Response{
			Error: "failed to create review app: unauthorized",
		}

		return &res, nil
	}

	res := proto.Response{
		Hostname: req.GetAppName() + req.GetBranch(),
		Error:    "",
	}

	return &res, err
}

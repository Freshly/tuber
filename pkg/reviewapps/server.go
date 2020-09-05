package reviewapps

import (
	"context"
	"fmt"
	"tuber/pkg/proto"
)

// Server is the ReviewApp GRPC service
type Server struct {
	proto.UnimplementedTuberServer
}

// CreateReviewApp creates a review app
func (s *Server) CreateReviewApp(ctx context.Context, in *proto.CreateReviewAppRequest) (*proto.CreateReviewAppResponse, error) {
	return &proto.CreateReviewAppResponse{
		Hostname: fmt.Sprintf("%s - %s\n%s", in.AppName, in.Branch, in.Token),
	}, nil
}

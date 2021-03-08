package reviewapps

import (
	"context"
	"fmt"
	"tuber/pkg/proto"

	"go.uber.org/zap"
)

// Server is the ReviewApp GRPC service
type Server struct {
	ClusterDefaultHost string
	ProjectName        string
	Credentials        []byte
	Logger             *zap.Logger
	proto.UnimplementedTuberServer
}

// CreateReviewApp creates a review app
func (s *Server) CreateReviewApp(ctx context.Context, in *proto.CreateReviewAppRequest) (*proto.CreateReviewAppResponse, error) {
	permitted, err := canCreate(s.Logger, in.AppName, in.Token)
	if err != nil {
		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}
	if !permitted {
		return &proto.CreateReviewAppResponse{
			Error: "not permitted to create a review app from " + in.AppName,
		}, nil
	}

	appName, err := CreateReviewApp(ctx, s.Logger, in.Branch, in.AppName, s.Credentials, s.ProjectName)

	if err != nil {
		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}

	var host string
	if s.ClusterDefaultHost == "" {
		host = appName
	} else {
		host = fmt.Sprintf("https://%s.%s/", appName, s.ClusterDefaultHost)
	}

	return &proto.CreateReviewAppResponse{
		Hostname: host,
	}, nil
}

func (s *Server) DeleteReviewApp(ctx context.Context, in *proto.DeleteReviewAppRequest) (*proto.DeleteReviewAppResponse, error) {
	logger := s.Logger.With(
		zap.String("appName", in.AppName),
	)

	err := DeleteReviewApp(ctx, in.AppName, s.Credentials, s.ProjectName)

	if err != nil {
		logger.Error("error deleting review app " + in.AppName + ": " + err.Error())
		return &proto.DeleteReviewAppResponse{Error: err.Error()}, nil
	}

	logger.Info("deleted review app: " + in.AppName)
	return &proto.DeleteReviewAppResponse{}, nil
}

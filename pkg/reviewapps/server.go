package reviewapps

import (
	"context"
	"fmt"
	"net/url"
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
	err := CreateReviewApp(ctx, s.Logger, in.Branch, in.AppName, in.Token, s.Credentials, s.ProjectName)

	if err != nil {
		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.CreateReviewAppResponse{
		Hostname: fmt.Sprintf("https://%s.%s/", reviewAppName, s.ClusterDefaultHost),
	}, nil
}

func (s *Server) DeleteReviewApp(ctx context.Context, in *proto.DeleteReviewAppRequest) (*proto.DeleteReviewAppResponse, error) {
	reviewAppName := in.GetAppName()

	logger := s.Logger.With(
		zap.String("appName", in.AppName),
	)

	err := DeleteReviewApp(ctx, reviewAppName, s.Credentials, s.ProjectName)

	if err != nil {
		logger.Error("error deleting review app " + reviewAppName + ": " + err.Error())
		return &proto.DeleteReviewAppResponse{Error: err.Error()}, nil
	}

	logger.Info("deleted review app: " + reviewAppName)
	return &proto.DeleteReviewAppResponse{}, nil
}

func reviewAppName(appName, branch string) string {
	return fmt.Sprintf("%s-%s", url.QueryEscape(appName), url.QueryEscape(branch))
}

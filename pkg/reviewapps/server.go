package reviewapps

import (
	"context"
	"fmt"
	"tuber/pkg/core"
	"tuber/pkg/proto"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Server is the ReviewApp GRPC service
type Server struct {
	ReviewAppsEnabled  bool
	ClusterDefaultHost string
	ProjectName        string
	Credentials        []byte
	Logger             *zap.Logger
	proto.UnimplementedTuberServer
}

// CreateReviewApp creates a review app
func (s *Server) CreateReviewApp(ctx context.Context, in *proto.CreateReviewAppRequest) (*proto.CreateReviewAppResponse, error) {
	if s.ReviewAppsEnabled == false {
		return &proto.CreateReviewAppResponse{
			Error: "review apps are not enabled for this cluster",
		}, nil
	}

	reviewAppName := reviewAppName(in.AppName, in.Branch)

	logger := s.Logger.With(
		zap.String("appName", in.AppName),
		zap.String("reviewAppName", reviewAppName),
		zap.String("branch", in.Branch),
	)

	logger.Info("checking permissions")
	if !canCreate(in.AppName, in.Token) {
		return &proto.CreateReviewAppResponse{
			Error: "not permitted to create a review app",
		}, nil
	}

	logger.Info("creating review app resources")
	err := NewReviewAppSetup(in.AppName, reviewAppName)
	if err != nil {
		logger.Info("error creating review app resources; tearing down")
		teardownErr := core.DestroyTuberApp(reviewAppName)
		if teardownErr != nil {
			logger.Info("error tearing down review app resources")
			return nil, teardownErr
		}

		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}

	logger.Info("creating and running review app trigger")
	removeTrigger, err := CreateAndRunTrigger(ctx, s.Credentials, in.AppName, s.ProjectName, reviewAppName, in.Branch)
	if err != nil {
		logger.Error("error creating trigger; no trigger resource created")

		if removeTrigger != nil {
			logger.Error("error creating trigger: removing trigger resources")

			rmvErr := removeTrigger()
			if rmvErr != nil {
				return nil, rmvErr
			}
		}

		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.CreateReviewAppResponse{
		Hostname: fmt.Sprintf("%s/%s", s.ClusterDefaultHost, reviewAppName),
	}, nil
}

func reviewAppName(appName, branch string) string {
	randStr := uuid.New().String()[0:8]

	if len(branch) > 8 {
		branch = branch[0:8]
	}

	if len(appName) > 8 {
		appName = appName[0:8]
	}

	return fmt.Sprintf("%s-%s-%s", appName, branch, randStr)
}

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

	logger.Info("creating review app")

	logger.Info("checking permissions")
	permitted, err := canCreate(logger, in.AppName, in.Token)
	if err != nil {
		return nil, err
	}

	if !permitted {
		return &proto.CreateReviewAppResponse{
			Error: "not permitted to create a review app",
		}, nil
	}

	sourceApp, err := core.FindApp(in.AppName)
	if err != nil {
		return &proto.CreateReviewAppResponse{
			Error: fmt.Sprintf("can't find source app. is %s managed by tuber?", in.AppName),
		}, nil
	}

	logger.Info("creating review app resources")

	err = NewReviewAppSetup(in.AppName, reviewAppName)
	if err != nil {
		logger.Error("error creating review app resources; tearing down", zap.Error(err))

		teardownErr := core.DestroyTuberApp(reviewAppName)
		if teardownErr != nil {
			logger.Error("error tearing down review app resources", zap.Error(teardownErr))
			return nil, teardownErr
		}

		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}

	logger.Info("creating app entry for review app")

	err = core.AddReviewAppConfig(reviewAppName, sourceApp.Repo, in.Branch)
	if err != nil {
		teardownErr := core.DestroyTuberApp(reviewAppName)
		if teardownErr != nil {
			logger.Error("error tearing down review app resources", zap.Error(teardownErr))
			return nil, teardownErr
		}

		return nil, err
	}

	logger.Info("creating and running review app trigger")

	removeTrigger, err := CreateAndRunTrigger(ctx, s.Credentials, sourceApp.Repo, s.ProjectName, reviewAppName, in.Branch)
	if err != nil {
		logger.Error("error creating trigger; no trigger resource created", zap.Error(err))

		teardownErr := core.DestroyTuberApp(reviewAppName)
		cleanupConfigErr := core.RemoveReviewAppConfig(reviewAppName)

		if removeTrigger != nil {
			logger.Error("error creating trigger: removing trigger resources")

			rmvErr := removeTrigger()
			if rmvErr != nil {
				return nil, rmvErr
			}
		}

		if teardownErr != nil {
			logger.Error("error tearing down review app resources", zap.Error(teardownErr))
			return nil, teardownErr
		}

		if cleanupConfigErr != nil {
			logger.Error("error removing config entry for app", zap.Error(cleanupConfigErr))
			return nil, cleanupConfigErr
		}

		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.CreateReviewAppResponse{
		Hostname: fmt.Sprintf("https://%s.%s/", reviewAppName, s.ClusterDefaultHost),
	}, nil
}

func (s *Server) DeleteReviewApp(ctx context.Context, in *proto.DeleteReviewAppRequest) (*proto.DeleteReviewAppResponse, error) {
	res := &proto.DeleteReviewAppResponse{}
	return res, nil
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

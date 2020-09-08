package reviewapps

import (
	"context"
	"fmt"
	"tuber/pkg/core"
	"tuber/pkg/proto"

	"github.com/google/uuid"
)

// Server is the ReviewApp GRPC service
type Server struct {
	ReviewAppsEnabled  bool
	ClusterDefaultHost string
	ProjectName        string
	Credentials        []byte
	proto.UnimplementedTuberServer
}

// CreateReviewApp creates a review app
func (s *Server) CreateReviewApp(ctx context.Context, in *proto.CreateReviewAppRequest) (*proto.CreateReviewAppResponse, error) {
	if s.ReviewAppsEnabled == false {
		return &proto.CreateReviewAppResponse{
			Error: "review apps are not enabled for this cluster",
		}, nil
	}

	if !canCreate(in.AppName, in.Token) {
		return &proto.CreateReviewAppResponse{
			Error: "not permitted to create a review app",
		}, nil
	}

	reviewAppName := reviewAppName(in.AppName, in.Branch)

	err := NewReviewAppSetup(in.AppName, reviewAppName)
	if err != nil {
		teardownErr := core.DestroyTuberApp(reviewAppName)
		if teardownErr != nil {
			return nil, teardownErr
		}

		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}

	removeTrigger, err := CreateAndRunTrigger(ctx, s.Credentials, in.AppName, s.ProjectName, reviewAppName, in.Branch)
	if err != nil {
		if removeTrigger != nil {
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
	appName = appName[0:8]
	branch = branch[0:8]
	randStr := uuid.New().String()[0:8]

	return fmt.Sprintf("%s-%s-%s", appName, branch, randStr)
}

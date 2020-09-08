package reviewapps

import (
	"context"
	"fmt"
	"tuber/pkg/proto"

	"github.com/google/uuid"
)

// Server is the ReviewApp GRPC service
type Server struct {
	proto.UnimplementedTuberServer
}

// CreateReviewApp creates a review app
func (s *Server) CreateReviewApp(ctx context.Context, in *proto.CreateReviewAppRequest) (*proto.CreateReviewAppResponse, error) {
	if !canCreate(in.AppName, in.Token) {
		return &proto.CreateReviewAppResponse{
			Error: "not permitted to create a review app",
		}, nil
	}

	reviewAppName := reviewAppName()

	err := NewReviewAppSetup(in.AppName, reviewAppName)
	if err != nil {
		teardownErr := ReviewAppTearDown(in.AppName)
		if teardownErr != nil {
			return &proto.CreateReviewAppResponse{
				Error: teardownErr.Error(),
			}, nil
		}

		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}

	// TODO: Where can I get the project name?
	removeTrigger, err := CreateAndRunTrigger(ctx, []byte(""), in.AppName, "<PROJECT>", reviewAppName, in.Branch)
	if err != nil {
		if removeTrigger != nil {
			rmvErr := removeTrigger()
			if rmvErr != nil {
				return &proto.CreateReviewAppResponse{
					Error: err.Error(),
				}, nil
			}

			return &proto.CreateReviewAppResponse{
				Error: err.Error(),
			}, nil
		}
	}

	return &proto.CreateReviewAppResponse{
		Hostname: fmt.Sprintf("%s - %s\n%s", in.AppName, in.Branch, in.Token),
	}, nil
}

func reviewAppName() string {
	return uuid.New().String()[0:8]
}

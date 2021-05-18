package reviewapps

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/pkg/db"
	"github.com/freshly/tuber/pkg/proto"

	"go.uber.org/zap"
)

// Server is the ReviewApp GRPC service
type Server struct {
	clusterDefaultHost string
	projectName        string
	credentials        []byte
	logger             *zap.Logger
	db                 *db.DB
	proto.UnimplementedTuberServer
}

func NewServer(logger *zap.Logger, creds []byte, db *db.DB, clusterDefaultHost string, projectName string) *Server {
	return &Server{
		clusterDefaultHost: clusterDefaultHost,
		projectName:        projectName,
		credentials:        creds,
		logger:             logger,
		db:                 db,
	}
}

// CreateReviewApp creates a review app
func (s *Server) CreateReviewApp(ctx context.Context, in *proto.CreateReviewAppRequest) (*proto.CreateReviewAppResponse, error) {
	permitted, err := canCreate(s.logger, s.db, in.AppName, in.Token)
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

	appName, err := CreateReviewApp(ctx, s.db, s.logger, in.Branch, in.AppName, s.credentials, s.projectName)

	if err != nil {
		return &proto.CreateReviewAppResponse{
			Error: err.Error(),
		}, nil
	}

	var host string
	if s.clusterDefaultHost == "" {
		host = appName
	} else {
		host = fmt.Sprintf("https://%s.%s/", appName, s.clusterDefaultHost)
	}

	return &proto.CreateReviewAppResponse{
		Hostname: host,
	}, nil
}

func (s *Server) DeleteReviewApp(ctx context.Context, in *proto.DeleteReviewAppRequest) (*proto.DeleteReviewAppResponse, error) {
	logger := s.logger.With(
		zap.String("appName", in.AppName),
	)

	err := DeleteReviewApp(ctx, s.db, in.AppName, s.credentials, s.projectName)

	if err != nil {
		logger.Error("error deleting review app " + in.AppName + ": " + err.Error())
		return &proto.DeleteReviewAppResponse{Error: err.Error()}, nil
	}

	logger.Info("deleted review app: " + in.AppName)
	return &proto.DeleteReviewAppResponse{}, nil
}

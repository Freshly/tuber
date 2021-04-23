package adminserver

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"
)

type server struct {
	projectName         string
	reviewAppsEnabled   bool
	cloudbuildClient    *cloudbuild.Service
	clusterDefaultHost  string
	triggersProjectName string
	logger              *zap.Logger
	creds               []byte
}

func Start(ctx context.Context, logger *zap.Logger, triggersProjectName string, creds []byte, reviewAppsEnabled bool, clusterDefaultHost string) error {
	var cloudbuildClient *cloudbuild.Service

	if reviewAppsEnabled {
		cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(creds))
		if err != nil {
			return err
		}
		cloudbuildClient = cloudbuildService
	}

	return server{
		projectName:         triggersProjectName,
		reviewAppsEnabled:   reviewAppsEnabled,
		cloudbuildClient:    cloudbuildClient,
		clusterDefaultHost:  clusterDefaultHost,
		triggersProjectName: triggersProjectName,
		logger:              logger,
		creds:               creds,
	}.start()
}

func (s server) start() error {
	router := gin.Default()
	router.LoadHTMLGlob("pkg/adminserver/templates/*")
	tuber := router.Group("/tuber")
	{
		tuber.GET("/", s.dashboard)
	}
	apps := tuber.Group("/apps")
	{
		apps.GET("/:appName", s.app)
		apps.GET("/:appName/reviewapps/:reviewAppName", s.reviewApp)
		apps.GET("/:appName/reviewapps/:reviewAppName/delete", s.deleteReviewApp)
		apps.POST("/:appName/createReviewApp", s.createReviewApp)
	}
	router.Run(":3000")
	return nil
}

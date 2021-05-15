package adminserver

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/pkg/core"
	"github.com/gin-contrib/cors"
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
	db                  *core.Data
	port                string
}

func Start(ctx context.Context, logger *zap.Logger, db *core.Data, triggersProjectName string, creds []byte, reviewAppsEnabled bool, clusterDefaultHost string, port string) error {
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
		db:                  db,
		port:                port,
	}.start()
}

func (s server) start() error {
	var err error

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3002"},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.LoadHTMLGlob("pkg/adminserver/templates/*")

	tuber := router.Group("/tuber")
	{
		tuber.GET("/", s.dashboard)

		tuber.Any("/graphql", gin.WrapH(graph.Handler(s.db)))
		tuber.GET("/graphql/playground", gin.WrapF(playground.Handler("GraphQL playground", "/tuber/graphql")))

		apps := tuber.Group("/apps")
		{
			apps.GET("/:appName", s.app)
			apps.GET("/:appName/reviewapps/:reviewAppName", s.reviewApp)
			apps.GET("/:appName/reviewapps/:reviewAppName/delete", s.deleteReviewApp)
			apps.POST("/:appName/createReviewApp", s.createReviewApp)
		}
	}

	if s.port == "" {
		err = router.Run(":3000")
	} else {
		err = router.Run(fmt.Sprintf(":%s", s.port))
	}

	return err
}

package adminserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"
)

type handler interface {
	path() string
	handle(http.ResponseWriter, *http.Request)
}

type server struct {
	projectName       string
	reviewAppsEnabled bool
	handlers          []handler
	cloudbuildClient  *cloudbuild.Service
}

type createReviewAppsArgs struct {
	ctx    context.Context
	logger *zap.Logger
	creds  []byte
}

func Start(ctx context.Context, logger *zap.Logger, triggersProjectName string, creds []byte, reviewAppsEnabled bool, clusterDefaultHost string) error {
	var cloudbuildClient *cloudbuild.Service
	var reviewAppsArgs *createReviewAppsArgs
	if reviewAppsEnabled {
		cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(creds))
		if err != nil {
			return err
		}
		cloudbuildClient = cloudbuildService

		reviewAppsArgs = &createReviewAppsArgs{
			ctx:    ctx,
			logger: logger,
			creds:  creds,
		}
	}
	return server{
		projectName:       triggersProjectName,
		reviewAppsEnabled: reviewAppsEnabled,
		handlers: []handler{
			dashboardHandler(),
			sourceAppHandler().setup(reviewAppsEnabled, clusterDefaultHost),
			reviewAppHandler().setup(reviewAppsEnabled, cloudbuildClient, triggersProjectName, clusterDefaultHost),
			createReviewAppHandler().setup(reviewAppsEnabled, reviewAppsArgs, triggersProjectName),
		},
	}.start()
}

func (s server) start() error {
	r := mux.NewRouter().PathPrefix("/tuber").Subrouter().StrictSlash(true)
	for _, handler := range s.handlers {
		r.HandleFunc(handler.path(), handler.handle)
	}
	http.Handle("/", r)
	return http.ListenAndServe(":3000", nil)
}

func dumbLink(appname string, clusterDefaultHost string) string {
	return fmt.Sprintf("https://%s.%s/", appname, clusterDefaultHost)
}

package adminserver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/events"
	"github.com/go-http-utils/logger"
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
	db                  *core.DB
	port                string
	processor           *events.Processor
	clusterName         string
	clusterRegion       string
	prefix              string
	useDevServer        bool
}

func Start(ctx context.Context, logger *zap.Logger, db *core.DB, processor *events.Processor, triggersProjectName string, creds []byte, reviewAppsEnabled bool, clusterDefaultHost string, port string, clusterName string, clusterRegion string, prefix string, useDevServer bool) error {
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
		processor:           processor,
		clusterName:         clusterName,
		clusterRegion:       clusterRegion,
		prefix:              prefix,
		useDevServer:        useDevServer,
	}.start()
}

func localDevServer(res http.ResponseWriter, req *http.Request) {
	remote, err := url.Parse("http://localhost:3002")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(proxyReq *http.Request) {
		proxyReq.Header = req.Header
		proxyReq.Host = remote.Host
		proxyReq.URL.Scheme = remote.Scheme
		proxyReq.URL.Host = remote.Host
		proxyReq.URL.Path = req.URL.Path
	}

	proxy.ServeHTTP(res, req)
}

func (s server) prefixed(route string) string {
	return fmt.Sprintf("%s%s", s.prefix, route)
}

func (s server) start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.prefixed("/graphql/playground"), playground.Handler("GraphQL playground", s.prefixed("/graphql")))
	mux.Handle(s.prefixed("/graphql"), graph.Handler(s.db, s.processor, s.logger, s.creds, s.triggersProjectName, s.clusterName, s.clusterRegion, s.reviewAppsEnabled))

	if s.useDevServer {
		mux.HandleFunc(s.prefixed("/"), localDevServer)
	}

	handler := logger.Handler(mux, os.Stdout, logger.DevLoggerType)

	port := ":3000"
	if s.port != "" {
		port = fmt.Sprintf(":%s", s.port)
	}

	return http.ListenAndServe(port, handler)
}

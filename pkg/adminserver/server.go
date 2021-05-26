package adminserver

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/pkg/core"
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
}

func Start(ctx context.Context, logger *zap.Logger, db *core.DB, triggersProjectName string, creds []byte, reviewAppsEnabled bool, clusterDefaultHost string, port string) error {
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

//go:embed web/out/* web/out/_next/static/chunks/pages/* web/out/_next/static/NEXT_MAGIC_FOLDER_REPLACE/*
var staticFiles embed.FS

func fixpath(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// log.Println("hello path", r.URL.Path)
		r.URL.Path = strings.Replace(r.URL.Path, "/localtunnel", "/web/out", 1)
		r.URL.RawPath = strings.Replace(r.URL.RawPath, "/localtunnel", "/web/out", 1)

		log.Println("path", r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

func (s server) start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/localtunnel/graphql/playground", playground.Handler("GraphQL playground", "/tuber/graphql"))
	mux.Handle("/localtunnel/graphql", graph.Handler(s.db, s.logger, s.creds, s.triggersProjectName))

	if false {
		mux.HandleFunc("/localtunnel/", localDevServer)
	} else {
		var staticFS = http.FS(staticFiles)
		fs := http.FileServer(staticFS)
		mux.HandleFunc("/localtunnel/", fixpath(fs))
	}

	handler := logger.Handler(mux, os.Stdout, logger.DevLoggerType)

	port := ":3000"
	if s.port != "" {
		port = fmt.Sprintf(":%s", s.port)
	}

	return http.ListenAndServe(port, handler)
}

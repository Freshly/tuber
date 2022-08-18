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
	"github.com/gorilla/securecookie"
	"go.uber.org/zap"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"
)

type server struct {
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
	secureCookie        *securecookie.SecureCookie
}

func Start(ctx context.Context, logger *zap.Logger, db *core.DB, processor *events.Processor, triggersProjectName string,
	creds []byte, reviewAppsEnabled bool, clusterDefaultHost string, port string, clusterName string, clusterRegion string,
	prefix string, useDevServer bool, secureCookie *securecookie.SecureCookie) error {
	var cloudbuildClient *cloudbuild.Service

	if reviewAppsEnabled {
		cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(creds))
		if err != nil {
			return err
		}
		cloudbuildClient = cloudbuildService
	}

	return server{
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
		secureCookie:        secureCookie,
	}.start()
}

func (s server) prefixed(route string) string {
	return fmt.Sprintf("%s%s", s.prefix, route)
}

func (s server) receiveAuthRedirect(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received auth redirect")
	http.Redirect(w, r, "/tuber/", http.StatusFound)
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h2>unauthorized: "+r.URL.Query().Get("error")+"</h2>")
}

func (s server) start() error {
	var proxyUrl = "http://tuber-frontend.tuber-frontend.svc.cluster.local:3000"
	if s.useDevServer {
		proxyUrl = "http://localhost:3002"
	}
	remote, err := url.Parse(proxyUrl)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	mux := http.NewServeMux()
	mux.HandleFunc(s.prefixed("/"), func(w http.ResponseWriter, r *http.Request) { proxy.ServeHTTP(w, r) })
	mux.HandleFunc(s.prefixed("/_next/"), func(w http.ResponseWriter, r *http.Request) { proxy.ServeHTTP(w, r) })
	mux.HandleFunc(s.prefixed("/graphql/playground"), playground.Handler("GraphQL playground", s.prefixed("/graphql")))
	mux.Handle(s.prefixed("/graphql"), graph.Handler(s.db, s.processor, s.logger, s.creds, s.triggersProjectName, s.clusterName, s.clusterRegion, s.reviewAppsEnabled))
	mux.HandleFunc(s.prefixed("/unauthorized/"), unauthorized)
	mux.HandleFunc(s.prefixed("/auth/"), s.receiveAuthRedirect)

	handler := logger.Handler(mux, os.Stdout, logger.DevLoggerType)

	port := ":3000"
	if s.port != "" {
		port = fmt.Sprintf(":%s", s.port)
	}

	if s.useDevServer {
		fmt.Println("listening on: http://localhost:" + s.port + s.prefix)
	}

	return http.ListenAndServe(port, handler)
}

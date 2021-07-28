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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

func debugTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header)
		fmt.Println(r.Cookies())
		next.ServeHTTP(w, r)
	})
}

var changeMeToEnvLater = "asdfasdf"

func lauren(w http.ResponseWriter, r *http.Request) {
	fmt.Println("yeah use this ya idiot 1060298202659-ulji10nd13lpp7ltldhko6j3fq9ub8i9.apps.googleusercontent.com")
	c := &oauth2.Config{
		RedirectURL:  "https://admin.freshlyhq.com/tuber/auth",
		ClientID:     "1060298202659-ulji10nd13lpp7ltldhko6j3fq9ub8i9.apps.googleusercontent.com",
		ClientSecret: "-VpmGDw5xcc-SEbZUlAgYx1A",
		Scopes:       []string{"openid", "email", "https://www.googleapis.com/auth/cloud-platform"},
		Endpoint:     google.Endpoint,
	}
	http.Redirect(w, r, c.AuthCodeURL(changeMeToEnvLater), 301)
}

func receiveAuthRedirect(w http.ResponseWriter, r *http.Request) {
	queryVals := r.URL.Query()
	if queryVals.Get("error") != "" {
		fmt.Fprintf(w, fmt.Sprintf("<h1>error in the redirect response</h1><h1>%s</h1>", queryVals.Get("error")))
		return
	}
	if queryVals.Get("code") == "" {
		fmt.Fprintf(w, "<h1>no code or error?</h1>")
		return
	}
	c := &oauth2.Config{
		RedirectURL:  "https://admin.freshlyhq.com/tuber/auth",
		ClientID:     "1060298202659-ulji10nd13lpp7ltldhko6j3fq9ub8i9.apps.googleusercontent.com",
		ClientSecret: "-VpmGDw5xcc-SEbZUlAgYx1A",
		Scopes:       []string{"openid", "email", "https://www.googleapis.com/auth/cloud-platform"},
		Endpoint:     google.Endpoint,
	}
	token, err := c.Exchange(context.Background(), queryVals.Get("code"))
	if err != nil {
		fmt.Fprintf(w, fmt.Sprintf("<h1>exchange error :(</h1><h1>%s</h1>", err.Error()))
		return
	}
	fmt.Fprintf(w, fmt.Sprintf("<h1>success???????</h1><h1>%s</h1>", token))
}

func (s server) start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.prefixed("/graphql/playground"), playground.Handler("GraphQL playground", s.prefixed("/graphql")))
	mux.Handle(s.prefixed("/graphql"), debugTime(graph.Handler(s.db, s.processor, s.logger, s.creds, s.triggersProjectName, s.clusterName, s.clusterRegion, s.reviewAppsEnabled)))
	mux.HandleFunc(s.prefixed("/lauren/"), lauren)
	mux.HandleFunc(s.prefixed("/auth/"), receiveAuthRedirect)

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

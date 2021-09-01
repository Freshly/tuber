package adminserver

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/pkg/config"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/iap"
	"github.com/freshly/tuber/pkg/oauth"
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
	authenticator       *oauth.Authenticator
	secureCookie        *securecookie.SecureCookie
}

func Start(ctx context.Context, logger *zap.Logger, db *core.DB, processor *events.Processor, triggersProjectName string,
	creds []byte, reviewAppsEnabled bool, clusterDefaultHost string, port string, clusterName string, clusterRegion string,
	prefix string, useDevServer bool, authenticator *oauth.Authenticator, secureCookie *securecookie.SecureCookie) error {
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
		authenticator:       authenticator,
		secureCookie:        secureCookie,
	}.start()
}

func (s server) prefixed(route string) string {
	return fmt.Sprintf("%s%s", s.prefix, route)
}

func (s server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("requireAuth running")
		var err error
		if s.useDevServer {
			w, r = s.devServerAuth(w, r)
		}
		var authed bool
		r, authed = s.authenticator.TrySetHeaderAuthContext(r)
		if authed {
			next.ServeHTTP(w, r)
			return
		}

		w, r, authed, err = s.authenticator.TrySetCookieAuthContext(w, r, s.secureCookie)
		if err != nil {
			s.logger.Error(fmt.Sprintf("cookie auth error: %v", err.Error()))
			http.Redirect(w, r, s.authenticator.RefreshTokenConsentUrl(), http.StatusMovedPermanently)
			return
		}

		if authed {
			next.ServeHTTP(w, r)
			return
		}

		http.Redirect(w, r, s.authenticator.RefreshTokenConsentUrl(), http.StatusMovedPermanently)
	})
}

func (s server) receiveAuthRedirect(w http.ResponseWriter, r *http.Request) {
	queryVals := r.URL.Query()
	if queryVals.Get("error") != "" {
		http.Redirect(w, r, fmt.Sprintf("/tuber/unauthorized/&error=%s", queryVals.Get("error")), http.StatusUnauthorized)
		return
	}
	if queryVals.Get("code") == "" {
		http.Redirect(w, r, fmt.Sprintf("/tuber/unauthorized/&error=%s", "no auth code returned from iap"), http.StatusUnauthorized)
		return
	}
	cookies, err := s.authenticator.GetTokenCookiesFromAuthToken(r.Context(), queryVals.Get("code"), s.secureCookie)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/tuber/unauthorized/&error=%s", err.Error()), http.StatusUnauthorized)
		return
	}
	for _, cookie := range cookies {
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/tuber/", http.StatusMovedPermanently)
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
	mux.HandleFunc("/tuber/", func(w http.ResponseWriter, r *http.Request) { s.requireAuth(proxy).ServeHTTP(w, r) })
	mux.HandleFunc("/tuber/_next/", func(w http.ResponseWriter, r *http.Request) { proxy.ServeHTTP(w, r) })
	mux.HandleFunc(s.prefixed("/graphql/playground"), playground.Handler("GraphQL playground", s.prefixed("/graphql")))
	mux.Handle(s.prefixed("/graphql"), s.requireAuth(graph.Handler(s.db, s.processor, s.logger, s.creds, s.triggersProjectName, s.clusterName, s.clusterRegion, s.reviewAppsEnabled)))
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

// cmon it's kinda cool
func (s server) devServerAuth(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	var refreshFound bool
	var accessFound bool
	for _, cookie := range r.Cookies() {
		if cookie.Name == oauth.RefreshTokenCookieKey() && cookie.Value != "" {
			refreshFound = true
		}
		if cookie.Name == oauth.AccessTokenCookieKey() && cookie.Value != "" {
			accessFound = true
		}
	}
	if !refreshFound || !accessFound {
		c, err := config.Load()
		if err != nil {
			fmt.Println(err)
			return w, r
		}

		cluster, err := c.CurrentClusterConfig()
		if err != nil {
			fmt.Println(err)
			return w, r
		}
		tokens, err := iap.CreateIDToken(cluster.Auth.Audience)
		if err != nil {
			fmt.Println(err)
			return w, r
		}

		encodedRefresh, err := s.secureCookie.Encode(oauth.RefreshTokenCookieKey(), tokens.RefreshToken)
		if err != nil {
			fmt.Println(err)
			return w, r
		}

		encodedAccess, err := s.secureCookie.Encode(oauth.RefreshTokenCookieKey(), tokens.AccessToken)
		if err != nil {
			fmt.Println(err)
			return w, r
		}
		expires := int64(math.Round(tokens.Raw.Raw.(map[string]interface{})["expires_in"].(float64)))

		cookies := []*http.Cookie{
			{Name: oauth.RefreshTokenCookieKey(), Value: encodedRefresh, HttpOnly: true, Secure: true, Path: "/"},
			{Name: oauth.AccessTokenCookieKey(), Value: encodedAccess, HttpOnly: true, Secure: true, Path: "/", Expires: time.Now().Add(time.Minute * time.Duration(expires))},
			{Name: oauth.AccessTokenExpirationCookieKey(), Value: time.Now().Add(time.Minute * time.Duration(expires)).Format(time.RFC3339), HttpOnly: true, Secure: true, Path: "/", Expires: time.Now().Add(time.Minute * time.Duration(expires))},
		}
		for _, cookie := range cookies {
			http.SetCookie(w, cookie)
			r.AddCookie(cookie)
		}
	}
	return w, r
}

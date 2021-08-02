package graph

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/freshly/tuber/graph/generated"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/oauth"
	"go.uber.org/zap"
)

func Handler(db *core.DB, processor *events.Processor, logger *zap.Logger, credentials []byte, projectName string, clusterName string, clusterRegion string, reviewAppsEnabled bool, authenticator *oauth.Authenticator) http.Handler {
	return handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers: NewResolver(db, logger, processor, credentials, projectName, clusterName, clusterRegion, reviewAppsEnabled, authenticator),
			},
		),
	)
}

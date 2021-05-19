package graph

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/freshly/tuber/graph/generated"
	"github.com/freshly/tuber/pkg/core"
	"go.uber.org/zap"
)

func Handler(db *core.DB, logger *zap.Logger, credentials []byte, projectName string) http.Handler {
	return handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers: NewResolver(db, logger, credentials, projectName),
			},
		),
	)
}

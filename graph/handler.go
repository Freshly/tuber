package graph

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/freshly/tuber/graph/generated"
	"github.com/freshly/tuber/pkg/core"
)

func Handler(db *core.DB) http.Handler {
	return handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{
				Resolvers: NewResolver(db),
			},
		),
	)
}

package cmd

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/graph/generated"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(graphqlServerCmd)
}

var graphqlServerCmd = &cobra.Command{
	Use:     "graphql-server",
	Short:   "Start tuber's pub/sub server",
	Run:     graphqlServer,
	PreRunE: promptCurrentContext,
}

func graphqlServer(cmd *cobra.Command, args []string) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4040"
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	http.Handle("/graphql", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

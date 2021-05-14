package cmd

import (
	"log"
	"net/http"
	"net/http/httputil"
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

func logVerbose(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("def received a request")

		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			panic(err)
		}
		log.Println(string(requestDump))

		next.ServeHTTP(w, r)
	})
}

func graphqlServer(cmd *cobra.Command, args []string) {
	db, err := db()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4040"
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: graph.NewResolver(db)}))

	http.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	http.Handle("/graphql", logVerbose(srv))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

package client

import (
	"context"
	"log"
	"os"

	"github.com/machinebox/graphql"
)

var graphqlURL = os.Getenv("GRAPHQL_URL")

func init() {
	if graphqlURL == "" {
		panic("cannot use graphql package without setting env GRAPHQL_URL")
	}
}

type GraphqlClient struct {
	client *graphql.Client
}

func New() *GraphqlClient {
	client := graphql.NewClient(graphqlURL)
	client.Log = func(s string) { log.Println(s) }

	return &GraphqlClient{
		client: client,
	}
}

func (g *GraphqlClient) Query(ctx context.Context, gql string, target interface{}) error {
	req := graphql.NewRequest(gql)

	// set any variables
	// req.Var("key", "value")

	// set header fields
	req.Header.Set("Cache-Control", "no-cache")

	if err := g.client.Run(ctx, req, &target); err != nil {
		return err
	}

	return nil
}

package client

import (
	"context"
	"log"

	"github.com/machinebox/graphql"
)

type GraphqlClient struct {
	client *graphql.Client
}

func New(graphqlURL string) *GraphqlClient {
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

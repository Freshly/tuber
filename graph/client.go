package graph

import (
	"context"
	"log"

	"github.com/machinebox/graphql"
	"github.com/spf13/viper"
)

type GraphqlClient struct {
	client *graphql.Client
}

func NewClient(clusterURL string) *GraphqlClient {
	graphqlURL := viper.GetString("graphql-host")

	if graphqlURL == "" {
		graphqlURL = clusterURL + viper.GetString("prefix") + "/graphql"
	} else {
		graphqlURL = graphqlURL + viper.GetString("prefix") + "/graphql"
	}

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

func (g *GraphqlClient) Mutation(ctx context.Context, gql string, key *int, input interface{}, target interface{}) error {
	req := graphql.NewRequest(gql)

	if key != nil {
		req.Var("key", *key)
	}

	if input != nil {
		req.Var("input", input)
	}

	// set header fields
	req.Header.Set("Cache-Control", "no-cache")

	if err := g.client.Run(ctx, req, &target); err != nil {
		return err
	}

	return nil
}

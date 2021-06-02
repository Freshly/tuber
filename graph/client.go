package graph

import (
	"context"
	"log"

	"github.com/freshly/tuber/pkg/iap"
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

type callOption struct {
	vars map[string]string
}

type callOptionFunc func() callOption

func WithVar(key string, val string) callOptionFunc {
	return func() callOption {
		return callOption{
			vars: map[string]string{key: val},
		}
	}
}

func (g *GraphqlClient) Query(ctx context.Context, gql string, target interface{}, options ...callOptionFunc) error {
	req := graphql.NewRequest(gql)

	for _, option := range options {
		res := option()

		if res.vars != nil {
			for k, v := range res.vars {
				req.Var(k, v)
			}
		}
	}

	token, err := iap.CreateIDToken()
	if err != nil {
		return err
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+token)

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

	token, err := iap.CreateIDToken()
	if err != nil {
		return err
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+token)

	if err := g.client.Run(ctx, req, &target); err != nil {
		return err
	}

	return nil
}
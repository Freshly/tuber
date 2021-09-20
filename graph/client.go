package graph

import (
	"context"
	"log"

	"github.com/freshly/tuber/pkg/iap"
	"github.com/machinebox/graphql"
	"github.com/spf13/viper"
)

type GraphqlClient struct {
	client      *graphql.Client
	IAPAudience string
}

func NewClient(clusterURL string, IAPAudience string) *GraphqlClient {
	graphqlURL := viper.GetString("TUBER_GRAPHQL_HOST")

	viper.SetDefault("TUBER_ADMINSERVER_PREFIX", "/tuber")
	if graphqlURL == "" {
		graphqlURL = clusterURL + viper.GetString("TUBER_ADMINSERVER_PREFIX") + "/graphql"
	} else {
		graphqlURL = graphqlURL + viper.GetString("TUBER_ADMINSERVER_PREFIX") + "/graphql"
	}

	client := graphql.NewClient(graphqlURL)

	if viper.GetBool("TUBER_DEBUG") {
		client.Log = func(s string) { log.Println(s) }
	}

	return &GraphqlClient{
		client:      client,
		IAPAudience: IAPAudience,
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

	tokens, err := iap.CreateIDToken(g.IAPAudience)
	if err != nil {
		return err
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+tokens.IDToken)
	req.Header.Set("Tuber-Token", tokens.AccessToken)

	err = g.client.Run(ctx, req, &target)
	if err != nil {
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

	tokens, err := iap.CreateIDToken(g.IAPAudience)
	if err != nil {
		return err
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", "Bearer "+tokens.IDToken)
	req.Header.Set("Tuber-Token", tokens.AccessToken)

	if err := g.client.Run(ctx, req, &target); err != nil {
		return err
	}

	return nil
}

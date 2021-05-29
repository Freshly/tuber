package cmd

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/goccy/go-yaml"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env [set || unset || get || list || file]",
	Short: "manage an app's environment",
}

var envSetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "set [app] [key] [value]",
	Short:        "set an environment variable",
	RunE:         envSet,
	Args:         cobra.ExactArgs(3),
	PreRunE:      promptCurrentContext,
}

var envUnsetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "unset [app] [key]",
	Short:        "unset an environment variable",
	RunE:         envUnset,
	Args:         cobra.ExactArgs(2),
	PreRunE:      promptCurrentContext,
}

var envGetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "get [app] [key]",
	Short:        "display the value of an environment variable",
	Args:         cobra.ExactArgs(2),
	RunE:         envGet,
	PreRunE:      displayCurrentContext,
}

var fileCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "file [app] [local filepath]",
	Short:        "batch set environment variables based on the contents of a yaml file",
	Args:         cobra.ExactArgs(2),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		err := k8s.CreateEnvFromFile(appName, args[1])

		if err != nil {
			return err
		}

		return k8s.Restart("deployments", appName)
	},
}

var envListCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "list [app]",
	Short:        "display all environment variables",
	RunE:         envList,
	Args:         cobra.ExactArgs(1),
	PreRunE:      displayCurrentContext,
}

func envSet(cmd *cobra.Command, args []string) error {
	appName := args[0]
	key := args[1]
	value := args[2]

	graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)

	input := &model.SetTupleInput{
		Name:  appName,
		Key:   key,
		Value: value,
	}

	var respData struct {
		setAppEnv *model.TuberApp
	}

	gql := `
		mutation($input: SetTupleInput!) {
			setAppEnv(input: $input) {
				name
			}
		}
	`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func envUnset(cmd *cobra.Command, args []string) error {
	appName := args[0]
	key := args[1]

	graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)

	input := &model.SetTupleInput{
		Name: appName,
		Key:  key,
	}

	var respData struct {
		setAppEnv *model.TuberApp
	}

	gql := `
		mutation($input: SetTupleInput!) {
			unsetAppEnv(input: $input) {
				name
			}
		}
	`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func envGet(cmd *cobra.Command, args []string) error {
	appName := args[0]
	key := args[1]

	m, err := getAllEnvGraphqlQuery(appName)
	if err != nil {
		return err
	}

	fmt.Println(m[key])
	return nil
}

func envList(cmd *cobra.Command, args []string) error {
	appName := args[0]

	m, err := getAllEnvGraphqlQuery(appName)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	fmt.Print(string(data))
	return nil
}

func getAllEnvGraphqlQuery(appName string) (map[string]string, error) {
	graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)

	gql := `
		query($name: String!) {
			getApp(name: $name) {
				name

				env {
					key
					value
				}
			}
		}
	`

	var respData struct {
		GetApp *model.TuberApp
	}

	if err := graphql.Query(context.Background(), gql, &respData, graph.WithVar("name", appName)); err != nil {
		return nil, err
	}

	m := make(map[string]string)

	for _, tuple := range respData.GetApp.Env {
		m[tuple.Key] = tuple.Value
	}

	return m, nil
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(fileCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envListCmd)
}

package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"

	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
	mapName := fmt.Sprintf("%s-env", appName)

	logger, err := createLogger()
	if err != nil {
		return err
	}

	logger.Info("env: unset",
		zap.String("name", appName),
		zap.String("key", key),
		zap.String("action", "change_env"),
	)

	err = k8s.RemoveSecretEntry(mapName, appName, key)
	if err != nil {
		return err
	}

	return k8s.Restart("deployments", appName)
}

func envGet(cmd *cobra.Command, args []string) (err error) {
	appName := args[0]
	mapName := fmt.Sprintf("%s-env", appName)
	key := args[1]
	config, err := k8s.GetConfigResource(mapName, appName, "Secret")

	if err != nil {
		return
	}

	v := config.Data[key]
	decoded, err := base64.StdEncoding.DecodeString(v)

	if err != nil {
		return
	}

	fmt.Println(string(decoded))
	return
}

func envList(cmd *cobra.Command, args []string) error {
	appName := args[0]
	mapName := fmt.Sprintf("%s-env", appName)
	config, err := k8s.GetSecret(appName, mapName)
	if err != nil {
		return err
	}

	var list []string
	for k, v := range config.Data {
		list = append(list, k+`: "`+v+`"`)
	}

	sort.Strings(list)
	for _, v := range list {
		fmt.Println(v)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(fileCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envListCmd)
}

package cmd

import (
	"context"
	"encoding/json"
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
	Use:          "set [appName (deprecated, use --app or -a)] [key] [value]",
	Short:        "set an environment variable",
	RunE:         envSet,
	Args:         cobra.RangeArgs(2, 3),
	PreRunE:      promptCurrentContext,
}

var envUnsetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "unset [appName (deprecated, use --app or -a)] [key]",
	Short:        "unset an environment variable",
	RunE:         envUnset,
	Args:         cobra.RangeArgs(1, 2),
	PreRunE:      promptCurrentContext,
}

var envGetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "get [appName (deprecated, use --app or -a)] [key]",
	Short:        "display the value of an environment variable",
	Args:         cobra.RangeArgs(1, 2),
	RunE:         envGet,
	PreRunE:      displayCurrentContext,
}

var fileCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "file [appName (deprecated, use --app or -a)] [local filepath]",
	Short:        "batch set environment variables based on the contents of a yaml file",
	Args:         cobra.RangeArgs(1, 2),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		var appName string
		var filePath string
		if len(args) == 2 {
			fmt.Println("App name as the first argument to this command is DEPRECATED. Please specify with -a or --app.")
			appName = args[0]
			filePath = args[1]
		}
		if len(args) == 1 {
			if appNameFlag == "" {
				return fmt.Errorf("app name required, specify with -a or --app")
			}
			appName = appNameFlag
			filePath = args[0]
		}
		err := k8s.CreateEnvFromFile(appName, filePath)

		if err != nil {
			return err
		}

		return k8s.Restart("deployments", appName)
	},
}

var listFmtFlag string
var envListCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "list [appName (deprecated, use --app or -a)]",
	Short:        "display all environment variables",
	RunE:         envList,
	Args:         cobra.RangeArgs(0, 1),
	PreRunE:      displayCurrentContext,
}

func envSet(cmd *cobra.Command, args []string) error {
	var appName string
	var key string
	var value string
	if len(args) == 3 {
		fmt.Println("App name as the first argument to this command is DEPRECATED. Please specify with -a or --app.")
		appName = args[0]
		key = args[1]
		value = args[2]
	}
	if len(args) == 2 {
		if appNameFlag == "" {
			return fmt.Errorf("app name required, specify with -a or --app")
		}
		appName = appNameFlag
		key = args[0]
		value = args[1]
	}

	graphql, err := gqlClient()
	if err != nil {
		return err
	}

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
	var appName string
	var key string
	if len(args) == 2 {
		fmt.Println("App name as the first argument to this command is DEPRECATED. Please specify with -a or --app.")
		appName = args[0]
		key = args[1]
	}
	if len(args) == 1 {
		if appNameFlag == "" {
			return fmt.Errorf("app name required, specify with -a or --app")
		}
		appName = appNameFlag
		key = args[0]
	}

	graphql, err := gqlClient()
	if err != nil {
		return err
	}

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
	var appName string
	var key string
	if len(args) == 2 {
		fmt.Println("App name as the first argument to this command is DEPRECATED. Please specify with -a or --app.")
		appName = args[0]
		key = args[1]
	}
	if len(args) == 1 {
		if appNameFlag == "" {
			return fmt.Errorf("app name required, specify with -a or --app")
		}
		appName = appNameFlag
		key = args[0]
	}

	m, err := getAllEnvGraphqlQuery(appName)
	if err != nil {
		return err
	}

	fmt.Println(m[key])
	return nil
}

func envList(cmd *cobra.Command, args []string) error {
	var appName string
	if len(args) == 1 {
		fmt.Println("App name as the first argument to this command is DEPRECATED. Please specify with -a or --app.")
		appName = args[0]
	} else {
		if appNameFlag == "" {
			return fmt.Errorf("app name required, specify with -a or --app")
		}
		appName = appNameFlag
	}

	m, err := getAllEnvGraphqlQuery(appName)
	if err != nil {
		return err
	}

	var output []byte
	switch listFmtFlag {
	case "env":
		for k, v := range m {
			output = append(output, []byte(fmt.Sprintf("%s = \"%s\"\n", k, v))...)
		}
	case "json":
		output, err = json.Marshal(m)
		if err != nil {
			return err
		}
	default:
		output, err = yaml.Marshal(m)
		if err != nil {
			return err
		}

	}

	fmt.Print(string(output))
	return nil
}

func getAllEnvGraphqlQuery(appName string) (map[string]string, error) {
	graphql, err := gqlClient()
	if err != nil {
		return nil, err
	}

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

	var respData []*model.Tuple

	err = graphql.Query(context.Background(), gql, &respData, graph.WithVar("name", appName))

	if err != nil {
		return nil, err
	}

	m := make(map[string]string)

	for _, tuple := range respData {
		m[tuple.Key] = tuple.Value
	}

	return m, nil
}

func init() {
	rootCmd.AddCommand(envCmd)
	envSetCmd.Flags().StringVarP(&appNameFlag, "app", "a", "", "app name")
	envCmd.AddCommand(envSetCmd)
	envUnsetCmd.Flags().StringVarP(&appNameFlag, "app", "a", "", "app name")
	envCmd.AddCommand(envUnsetCmd)
	fileCmd.Flags().StringVarP(&appNameFlag, "app", "a", "", "app name")
	envCmd.AddCommand(fileCmd)
	envGetCmd.Flags().StringVarP(&appNameFlag, "app", "a", "", "app name")
	envCmd.AddCommand(envGetCmd)
	envListCmd.Flags().StringVarP(&appNameFlag, "app", "a", "", "app name")
	envListCmd.Flags().StringVarP(&listFmtFlag, "output", "o", "", "output format to display environment variables")
	envCmd.AddCommand(envListCmd)
}

package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsDestroyCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "destroy [app name]",
	Short:         "destroy an app from the current cluster",
	Args:          cobra.ExactArgs(1),
	PreRunE:       promptCurrentContext,
	RunE:          destroyApp,
}

func destroyApp(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]

	input := &model.AppInput{
		Name: appName,
	}

	var respData struct {
		destroyApp *model.TuberApp
	}

	gql := `
		mutation($input: AppInput!) {
			destroyApp(input: $input) {
				name
			}
		}
	`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsCmd.AddCommand(appsDestroyCmd)
}

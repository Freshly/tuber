package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsRemoveCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "remove [app name]",
	Short:         "fully disconnects tuber from an app, without affecting the app",
	Args:          cobra.ExactArgs(1),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsRemoveCmd,
}

func runAppsRemoveCmd(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]

	input := &model.AppInput{
		Name: appName,
	}

	var respData struct {
		destoryApp *model.TuberApp
	}

	gql := `
			mutation($input: AppInput!) {
				removeApp(input: $input) {
					name
				}
			}
		`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsCmd.AddCommand(appsRemoveCmd)
}

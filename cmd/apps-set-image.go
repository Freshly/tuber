package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetImageTagCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "set-tag [app name] [image tag]",
	Short:         "set the docker image tag to deploy the app from",
	Args:          cobra.ExactArgs(2),
	PreRunE:       promptCurrentContext,
	RunE:          runSetImageTagCmd,
}

func runSetImageTagCmd(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]
	imageTag := args[1]

	input := &model.AppInput{
		Name:     appName,
		ImageTag: &imageTag,
	}

	var respData struct {
		updateApp *model.TuberApp
	}

	gql := `
			mutation($input: AppInput!) {
				updateApp(input: $input) {
					name
				}
			}
		`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsSetCmd.AddCommand(appsSetImageTagCmd)
}

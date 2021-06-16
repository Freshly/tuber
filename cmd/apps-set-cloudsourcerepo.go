package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetCloudSourceRepoCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "cloud-source-repo [app name] [repo-name]",
	Short:         "update the cloud source repo tuber will use when creating triggers for review apps",
	Args:          cobra.ExactArgs(2),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetCloudSourceRepoCmd,
}

func runAppsSetCloudSourceRepoCmd(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]
	repo := args[1]

	input := &model.AppInput{
		Name:            appName,
		CloudSourceRepo: &repo,
	}

	var respData struct {
		destroyApp *model.TuberApp
	}

	gql := `
		mutation($input: AppInput!) {
			setCloudSourceRepo(input: $input) {
				name
			}
		}
	`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsSetCmd.AddCommand(appsSetCloudSourceRepoCmd)
}

package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetSlackChannelCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "slack-channel [app name] [channel]",
	Short:         "update the slack channel for notifications for app",
	Args:          cobra.ExactArgs(2),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetSlackChannelCmd,
}

func runAppsSetSlackChannelCmd(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]
	channel := args[1]

	input := &model.AppInput{
		Name:         appName,
		SlackChannel: &channel,
	}

	var respData struct {
		destroyApp *model.TuberApp
	}

	gql := `
		mutation($input: AppInput!) {
			setSlackChannel(input: $input) {
				name
			}
		}
	`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsSetCmd.AddCommand(appsSetSlackChannelCmd)
}

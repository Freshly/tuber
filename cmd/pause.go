package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"

	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "pause [app name]",
	Short:        "pause deploys for the specified app",
	Args:         cobra.ExactArgs(1),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		graphql, err := gqlClient()
		if err != nil {
			return err
		}

		b := true
		input := &model.AppInput{
			Name:   appName,
			Paused: &b,
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
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}

package cmd

import (
	"context"

	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "resume [app name]",
	Short:        "resume deploys for the specified app",
	Args:         cobra.ExactArgs(1),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)

		b := false
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
	rootCmd.AddCommand(resumeCmd)
}

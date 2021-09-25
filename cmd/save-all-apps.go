package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"

	"github.com/spf13/cobra"
)

var saveAllAppsCmd = &cobra.Command{
	SilenceUsage: true,
	Hidden:       true,
	Use:          "save-all-apps",
	Short:        "general migration tool - internal, hidden, but also, optimistically, always safe",
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

		var respData interface{}

		gql := `
			mutation {
				saveAllApps {}
			}
		`

		return graphql.Mutation(context.Background(), gql, nil, input, &respData)
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}

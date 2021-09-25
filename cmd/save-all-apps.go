package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var saveAllAppsCmd = &cobra.Command{
	SilenceUsage: true,
	Hidden:       true,
	Use:          "save-all-apps",
	Short:        "general migration tool - internal, hidden, but also, optimistically, always safe",
	Args:         cobra.NoArgs,
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		graphql, err := gqlClient()
		if err != nil {
			return err
		}
		var respData interface{}
		gql := `
			mutation {
				saveAllApps
			}
		`

		return graphql.Mutation(context.Background(), gql, nil, nil, &respData)
	},
}

func init() {
	rootCmd.AddCommand(saveAllAppsCmd)
}

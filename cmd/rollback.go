package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "rollback [app]",
	Short:         "immediately roll back an app",
	RunE:          runRollback,
	PreRunE:       promptCurrentContext,
	Args:          cobra.ExactArgs(1),
	Long: `immediately rolls back to the resources (and image) applied during the last successful release, without monitoring for success.
Can be used to abort a running release as well, as tuber's definition of 'last successful release' is not updated until a running release finishes successfully.`,
}

func runRollback(cmd *cobra.Command, args []string) error {
	appName := args[0]
	graphql, err := gqlClient()
	if err != nil {
		return err
	}
	gql := `
		mutation($input: AppInput!) {
			rollback(input: $input) {
				name
			}
		}
	`

	input := &model.AppInput{
		Name: appName,
	}

	var respData struct {
		rollback *model.TuberApp
	}

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

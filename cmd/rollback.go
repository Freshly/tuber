package cmd

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph"
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
	Long: `immediately rolls back to the resources (and image) applied during the last successful release, without monitoring for success.
Can be used to abort a running release as well, as tuber's definition of 'last successful release' is not updated until a running release finishes successfully.`,
}

func runRollback(cmd *cobra.Command, args []string) error {
	appName := args[0]
	graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)
	gql := `
		mutation($input: "%s") {
			rollback(input: $input) {
				name
			}
		}
	`

	var respData struct {
		rollback *model.TuberApp
	}

	return graphql.Mutation(context.Background(), fmt.Sprintf(gql, appName), nil, appName, &respData)
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

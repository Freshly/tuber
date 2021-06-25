package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetRacExcludeCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "exclude [app name] [resource kind] [resource name]",
	Short:         "exclude a resource from being deployed along with review apps created from this source app",
	Args:          cobra.ExactArgs(3),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetRacExclude,
}

func runAppsSetRacExclude(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]
	kind := args[1]
	name := args[2]

	input := &model.SetResourceInput{
		AppName: appName,
		Kind:    kind,
		Name:    name,
	}

	var respData struct {
		updateApp *model.TuberApp
	}

	gql := `
			mutation($input: SetResourceInput!) {
				setRacExclusion(input: $input) {
					name
				}
			}
		`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsSetReviewappsConfigCmd.AddCommand(appsSetRacExcludeCmd)
}

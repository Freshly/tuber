package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetExcludeCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "exclude [app name] [resource kind] [resource name]",
	Short:         "exclude a resource from being deployed",
	Args:          cobra.ExactArgs(3),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetExclude,
}

func runAppsSetExclude(cmd *cobra.Command, args []string) error {
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
		setExcludedResource *model.TuberApp
	}

	gql := `
			mutation($input: SetResourceInput!) {
				setExcludedResource(input: $input) {
					name
				}
			}
		`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsSetCmd.AddCommand(appsSetExcludeCmd)
}

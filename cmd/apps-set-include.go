package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetIncludeCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "include [app name] [resource kind] [resource name]",
	Short:         "delete an exclusion for a resource that would otherwise prevent it from being deployed along with this app",
	Args:          cobra.ExactArgs(3),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetInclude,
}

func runAppsSetInclude(cmd *cobra.Command, args []string) error {
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
		unsetExcludedResource *model.TuberApp
	}

	gql := `
			mutation($input: SetResourceInput!) {
				unsetExcludedResource(input: $input) {
					name
				}
			}
		`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsSetCmd.AddCommand(appsSetIncludeCmd)
}

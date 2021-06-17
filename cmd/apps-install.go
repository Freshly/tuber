package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var istioEnabled bool

var appsInstallCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "install [app name] [docker tag] [--istio=<true(default) || false>]",
	Short:         "install a new app in the current cluster",
	Args:          cobra.ExactArgs(2),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsInstallCmd,
}

func runAppsInstallCmd(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}
	appName := args[0]
	imageTag := args[1]

	input := &model.AppInput{
		IsIstio:  &istioEnabled,
		Name:     appName,
		ImageTag: &imageTag,
	}

	var respData struct {
		createApp []*model.TuberApp
	}

	gql := `
			mutation($input: AppInput!) {
				createApp(input: $input) {
					name
				}
			}
		`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsInstallCmd.Flags().BoolVar(&istioEnabled, "istio", true, "enable (default) or disable istio sidecar injection for a new app")
	appsCmd.AddCommand(appsInstallCmd)
}

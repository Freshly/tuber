package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetGithubRepoCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "github-url [app name] [url]",
	Short:         "update the github url associated with the app for comparing shas while deploying",
	Long: `this is used exclusively for release diffs, so the url should read like 'https://github.com/Freshly/repoName/compare/'
under the hood, it simply appends 'oldSha...newSha' to whatever's in this field if present.`,
	Args:    cobra.ExactArgs(2),
	PreRunE: promptCurrentContext,
	RunE:    runAppsSetGithubRepoCmd,
}

func runAppsSetGithubRepoCmd(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]
	url := args[1]

	input := &model.AppInput{
		Name:       appName,
		GithubRepo: &url,
	}

	var respData struct {
		destroyApp *model.TuberApp
	}

	gql := `
		mutation($input: AppInput!) {
			setGithubRepo(input: $input) {
				name
			}
		}
	`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsSetCmd.AddCommand(appsSetGithubRepoCmd)
}

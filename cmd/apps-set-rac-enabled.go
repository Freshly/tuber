package cmd

import (
	"context"
	"strconv"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetRacEnabledCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "enabled [app name] [true or false]",
	Short:         "enable or disable review apps from being created from this source app",
	Args:          cobra.ExactArgs(2),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetRacEnabled,
}

func runAppsSetRacEnabled(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]
	enabledString := args[1]
	enabled, err := strconv.ParseBool(enabledString)
	if err != nil {
		return err
	}

	input := &model.SetRacEnabledInput{
		Name:    appName,
		Enabled: enabled,
	}

	var respData struct {
		updateApp *model.TuberApp
	}

	gql := `
			mutation($input: SetRacEnabledInput!) {
				setRacEnabled(input: $input) {
					name
				}
			}
		`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	appsSetReviewappsConfigCmd.AddCommand(appsSetRacEnabledCmd)
}

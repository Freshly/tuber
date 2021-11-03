package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "export [app name]",
	Short:        "export editable and importable version of an app",
	Args:         cobra.ExactArgs(1),
	PreRunE:      displayCurrentContext,
	RunE:         runExportCmd,
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func runExportCmd(cmd *cobra.Command, args []string) error {
	appName := args[0]
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	gql := `
		query {
			getApp(name: "%s") {
				cloudSourceRepo
				imageTag
				name
				paused
				reviewApp
				githubRepo
				reviewAppsConfig{
					enabled
					vars {
						key
						value
					}
					excludedResources {
						kind
						name
					}
				}
				slackChannel
				excludedResources {
					kind
					name
				}
				triggerID
				vars {
					key
					value
				}
			}
		}
	`

	var respData struct {
		GetApp *model.TuberApp
	}

	err = graphql.Query(context.Background(), fmt.Sprintf(gql, appName), &respData)
	if err != nil {
		return err
	}

	if respData.GetApp == nil {
		return fmt.Errorf("error retrieving app")
	}

	app := respData.GetApp

	out, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(appName+".json", out, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

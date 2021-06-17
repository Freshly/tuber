package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var appsInfoCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "info",
	Short:         "display everything tuber knows about an app",
	Args:          cobra.ExactArgs(1),
	RunE:          runAppsInfoCmd,
	PreRunE:       displayCurrentContext,
}

func runAppsInfoCmd(cmd *cobra.Command, args []string) error {
	appName := args[0]
	app, err := getApp(appName)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.Append([]string{"Name", app.Name})
	table.Append([]string{"ImageTag", app.ImageTag})
	table.Append([]string{"Current Tags", strings.Join(app.CurrentTags, "\n")})
	if app.ReviewAppsConfig != nil {
		table.Append([]string{"Review Apps Enabled", strconv.FormatBool(app.ReviewAppsConfig.Enabled)})
	} else {
		table.Append([]string{"Review Apps Enabled", "false"})
	}
	var vars []string
	for _, tuple := range app.Vars {
		vars = append(vars, tuple.Key+": "+tuple.Value)
	}
	table.Append([]string{"Vars", strings.Join(vars, "\n")})
	table.Append([]string{"Paused", strconv.FormatBool(app.Paused)})
	table.Append([]string{"Is Review App", strconv.FormatBool(app.ReviewApp)})
	if !app.ReviewApp && app.ReviewAppsConfig != nil {
		table.Append([]string{"Review Apps Enabled", strconv.FormatBool(app.ReviewAppsConfig.Enabled)})
		var reviewVars []string
		for _, tuple := range app.ReviewAppsConfig.Vars {
			reviewVars = append(reviewVars, tuple.Key+": "+tuple.Value)
		}
		table.Append([]string{"Starting Review App Vars", strings.Join(reviewVars, "\n")})
	}

	// var excludedResources []string
	// for _, resource := range app.ExcludedResources {
	// 	excludedResources = append(excludedResources, resource.Kind+": "+resource.Name)
	// }
	// table.Append([]string{"Excluded Resources", strings.Join(reviewVars, "\n")})

	table.Append([]string{"Slack Channel", app.SlackChannel})
	if app.ReviewApp {
		table.Append([]string{"Name", app.SourceAppName})
	}

	if app.State != nil {
		var state []string
		for _, resource := range app.State.Current {
			state = append(state, resource.Kind+": "+resource.Name)
		}
		table.Append([]string{"State", strings.Join(state, "\n")})
	}

	table.Append([]string{"Cloud Source Repo (for triggers)", app.CloudSourceRepo})
	table.Append([]string{"Trigger Id", app.TriggerID})

	table.Render()

	return nil
}

func init() {
	appsCmd.AddCommand(appsInfoCmd)
}

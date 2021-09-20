package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var appsInfoJsonFlag bool

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

	if appsInfoJsonFlag {
		out, err := json.Marshal(app)
		if err != nil {
			return err
		}

		os.Stdout.Write(out)
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.Append([]string{"Name", app.Name})
	table.Append([]string{"ImageTag", app.ImageTag})
	table.Append([]string{"Current Tags", strings.Join(app.CurrentTags, "\n")})
	var vars []string
	for _, tuple := range app.Vars {
		value := tuple.Value
		if len(value) > 70 {

			value = value[:69] + fmt.Sprintf("\n%*s%v", len(tuple.Key)+3, " ", value[69:])
		}
		vars = append(vars, tuple.Key+": "+value)
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
		if len(reviewVars) != 0 {
			table.Append([]string{"Starting Review App Vars", strings.Join(reviewVars, "\n")})
		}
		var racExclusions []string
		for _, resource := range app.ReviewAppsConfig.ExcludedResources {
			racExclusions = append(racExclusions, resource.Kind+": "+resource.Name)
		}
		if len(racExclusions) != 0 {
			table.Append([]string{"Starting Review App Excluded Resources", strings.Join(racExclusions, "\n")})
		}
	}

	var excludedResources []string
	for _, resource := range app.ExcludedResources {
		excludedResources = append(excludedResources, resource.Kind+": "+resource.Name)
	}
	table.Append([]string{"Excluded Resources", strings.Join(excludedResources, "\n")})

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
	appsInfoCmd.Flags().BoolVar(&appsInfoJsonFlag, "json", false, "output as json")
	appsCmd.AddCommand(appsInfoCmd)
}

package cmd

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/graph/model"
	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps [command]",
	Short: "A root command for app configurating.",
}

var istioEnabled bool
var jsonOutput bool

var appsInstallCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "install [app name] [docker tag] [--istio=<true(default) || false>]",
	Short:        "install a new app in the current cluster",
	Args:         cobra.ExactArgs(2),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)
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
	},
}

var appsSetImageTagCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "set-tag [app name ] [image tag]",
	Short:        "set the docker image tag to deploy the app from",
	Args:         cobra.ExactArgs(2),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)

		appName := args[0]
		imageTag := args[1]

		input := &model.AppInput{
			Name:     appName,
			ImageTag: &imageTag,
		}

		var respData struct {
			updateApp *model.TuberApp
		}

		gql := `
			mutation($input: AppInput!) {
				updateApp(input: $input) {
					name
				}
			}
		`

		return graphql.Mutation(context.Background(), gql, nil, input, &respData)
	},
}

var appsRemoveCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "remove [app name]",
	Short:        "remove an app from the tuber-apps config map in the current cluster",
	Args:         cobra.ExactArgs(1),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)

		appName := args[0]

		input := &model.AppInput{
			Name: appName,
		}

		var respData struct {
			destoryApp *model.TuberApp
		}

		gql := `
			mutation($input: AppInput!) {
				removeApp(input: $input) {
					name
				}
			}
		`

		return graphql.Mutation(context.Background(), gql, nil, input, &respData)
	},
}

var appsDestroyCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "destroy [app name]",
	Short:        "destroy an app from the current cluster",
	Args:         cobra.ExactArgs(1),
	PreRunE:      promptCurrentContext,
	RunE:         destroyApp,
}

func destroyApp(cmd *cobra.Command, args []string) error {
	graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)

	appName := args[0]

	input := &model.AppInput{
		Name: appName,
	}

	var respData struct {
		destroyApp *model.TuberApp
	}

	gql := `
		mutation($input: AppInput!) {
			destroyApp(input: $input) {
				name
			}
		}
	`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

var appsListCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "list",
	Short:        "List tuberapps",
	PreRunE:      displayCurrentContext,
	RunE: func(*cobra.Command, []string) (err error) {
		graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)

		gql := `
			query {
				getApps {
					name
					imageTag
				}
			}
		`

		var respData struct {
			GetApps []*model.TuberApp
		}

		if err := graphql.Query(context.Background(), gql, &respData); err != nil {
			return err
		}

		apps := respData.GetApps

		sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })

		if jsonOutput {
			out, err := json.Marshal(apps)

			if err != nil {
				return err
			}

			os.Stdout.Write(out)

			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Image"})
		table.SetBorder(false)

		for _, app := range apps {
			table.Append([]string{app.Name, app.ImageTag})
		}

		table.Render()
		return nil
	},
}

var appsInfoCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "info",
	Short:         "display everything tuber knows about an app",
	Args:          cobra.ExactArgs(1),
	RunE:          runAppsInfoCmd,
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
	appsInstallCmd.Flags().BoolVar(&istioEnabled, "istio", true, "enable (default) or disable istio sidecar injection for a new app")
	appsListCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as json")
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsInfoCmd)
	appsCmd.AddCommand(appsInstallCmd)
	appsCmd.AddCommand(appsRemoveCmd)
	appsCmd.AddCommand(appsDestroyCmd)
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsSetImageTagCmd)
}

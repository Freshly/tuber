package cmd

import (
	"context"
	"encoding/json"
	"os"
	"sort"

	"github.com/freshly/tuber/graph/client"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"

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
	Use:          "install [app name] [image tag] [--istio=<true(default) || false>]",
	Short:        "install a new app in the current cluster",
	Args:         cobra.ExactArgs(2),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		tag := args[1]

		err := core.NewAppSetup(appName, istioEnabled)
		if err != nil {
			return err
		}

		model.TuberApp{
			CloudSourceRepo:  "",
			ImageTag:         tag,
			Name:             appName,
			Paused:           false,
			ReviewApp:        false,
			ReviewAppsConfig: &model.ReviewAppsConfig{},
			SlackChannel:     "",
			SourceAppName:    "",
			State:            nil,
			Tag:              "",
			TriggerID:        "",
			Vars:             nil,
		}

		return core.AddSourceAppConfig(appName, repo, tag)
	},
}

var appsSetBranchCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "set-branch [app name] [branch name]",
	Short:        "set the branch to deploy the app from",
	Args:         cobra.ExactArgs(2),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		branch := args[1]

		app, err := core.FindApp(appName)

		if err != nil {
			return err
		}

		return core.AddSourceAppConfig(appName, app.Repo, branch)
	},
}

var appsSetRepoCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "set-repo [app name] [docker repo]",
	Short:        "set the docker repo to listen to for changes",
	Args:         cobra.ExactArgs(2),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		repo := args[1]

		app, err := core.FindApp(appName)

		if err != nil {
			return err
		}

		return core.AddSourceAppConfig(appName, repo, app.Tag)
	},
}
var appsRemoveCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "remove [app name]",
	Short:        "remove an app from the tuber-apps config map in the current cluster",
	Args:         cobra.ExactArgs(1),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		return core.RemoveSourceAppConfig(appName)
	},
}

var appsDestroyCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "destroy [app name]",
	Short:        "destroy an app from the current cluster",
	Args:         cobra.ExactArgs(1),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		return core.DestroyTuberApp(appName)
	},
}

var appsListCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "list",
	Short:        "List tuberapps",
	RunE: func(*cobra.Command, []string) (err error) {
		graphqlURL := viper.GetString("graphql-url")
		if graphqlURL == "" {
			config, err := getTuberConfig()
			if err != nil {
				return err
			}
			graphqlURL = config.CurrentClusterConfig().URL + "/graphql"
		}

		graphql := client.New(graphqlURL)

		gql := `
			query {
				getApps {
					name
					imageTag
				}
			}
		`

		var respData struct {
			GetApps []model.TuberApp
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

		// client.
		return nil
	},
}

func init() {
	appsInstallCmd.Flags().BoolVar(&istioEnabled, "istio", true, "enable (default) or disable istio sidecar injection for a new app")
	appsListCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as json")
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsInstallCmd)
	appsCmd.AddCommand(appsRemoveCmd)
	appsCmd.AddCommand(appsDestroyCmd)
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsSetBranchCmd)
	appsCmd.AddCommand(appsSetRepoCmd)
}

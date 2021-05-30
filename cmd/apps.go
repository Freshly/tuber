package cmd

import (
	"context"
	"encoding/json"
	"os"
	"sort"

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
			IsIstio:  istioEnabled,
			Name:     appName,
			ImageTag: imageTag,
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
			ImageTag: imageTag,
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

func init() {
	appsInstallCmd.Flags().BoolVar(&istioEnabled, "istio", true, "enable (default) or disable istio sidecar injection for a new app")
	appsListCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as json")
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsInstallCmd)
	appsCmd.AddCommand(appsRemoveCmd)
	appsCmd.AddCommand(appsDestroyCmd)
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsSetImageTagCmd)
}

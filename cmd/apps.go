package cmd

import (
	"os"
	"sort"
	"tuber/pkg/core"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps [command]",
	Short: "A root command for app configurating.",
}

var appsInstallCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "install [app name] [docker repo] [deploy tag]",
	Short:        "install a new app in the current cluster",
	Args:         cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		repo := args[1]
		tag := args[2]

		return core.CreateTuberApp(appName, repo, tag)
	},
}

var appsSetBranchCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "set-branch [app name] [branch name]",
	Short:        "set the branch to deploy the app from",
	Args:         cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		branch := args[1]

		app, err := core.FindApp(appName)

		if err != nil {
			return err
		}

		return core.AddAppConfig(appName, app.Repo, branch)
	},
}

var appsSetRepoCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "set-repo [app name] [docker repo]",
	Short:        "set the docker repo to listen to for changes",
	Args:         cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		repo := args[1]

		app, err := core.FindApp(appName)

		if err != nil {
			return err
		}

		return core.CreateTuberApp(appName, repo, app.Tag)
	},
}
var appsRemoveCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "remove [app name]",
	Short:        "remove an app from the tuber-apps config map in the current cluster",
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		return core.RemoveAppConfig(appName)
	},
}

var appsDestroyCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "destroy [app name]",
	Short:        "destroy an app from the current cluster",
	Args:         cobra.ExactArgs(1),
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
		apps, err := core.TuberApps()

		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Image"})
		table.SetBorder(false)

		sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })

		for _, app := range apps {
			table.Append([]string{app.Name, app.ImageTag})
		}

		table.Render()
		return
	},
}

func init() {
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsInstallCmd)
	appsCmd.AddCommand(appsRemoveCmd)
	appsCmd.AddCommand(appsDestroyCmd)
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsSetBranchCmd)
	appsCmd.AddCommand(appsSetRepoCmd)
}

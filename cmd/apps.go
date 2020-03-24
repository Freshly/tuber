package cmd

import (
	"os"
	"tuber/pkg/core"
	"tuber/pkg/k8s"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps [command]",
	Short: "A root command for app configurating.",
}

var istioEnabled bool

var appsInstallCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "install [app name] [docker repo] [deploy tag] [--no-istio (optional flag)]",
	Short:        "install a new app in the current cluster",
	Args:         cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		repo := args[1]
		tag := args[2]
		existsAlready, err := k8s.Exists("namespace", appName, appName)
		if err != nil {
			return err
		}

		if existsAlready {
			return core.AddAppConfig(appName, repo, tag)
		}

		err = core.NewAppSetup(appName, istioEnabled)
		if err != nil {
			return err
		}

		return core.AddAppConfig(appName, repo, tag)
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

		for _, app := range apps {
			table.Append([]string{app.Name, app.ImageTag})
		}

		table.Render()
		return
	},
}

func init() {
	appsInstallCmd.Flags().BoolVar(&istioEnabled, "istio", true, "enable (default) or disable istio for a new app")
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsInstallCmd)
	appsCmd.AddCommand(appsRemoveCmd)
	appsCmd.AddCommand(appsDestroyCmd)
	appsCmd.AddCommand(appsListCmd)
}

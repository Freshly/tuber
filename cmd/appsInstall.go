package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

// appsInstallCmd represents the appsInstall command
var appsInstallCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "install [app name] [docker repo] [deploy tag]",
	Short:        "install a new app in the current cluster",
	Args:         cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		repo := args[1]
		tag := args[2]

		return core.InstallTuberApp(appName, repo, tag)
	},
}

func init() {
	appsCmd.AddCommand(appsInstallCmd)
}

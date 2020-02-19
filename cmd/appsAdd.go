package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

// appsAddCmd represents the appsAdd command
var appsAddCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "add [app name] [docker repo] [deploy tag]",
	Short:        "add an app to the tuber-apps config map in the current the cluster",
	Args:         cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		repo := args[1]
		tag := args[2]

		return core.AddAppConfig(appName, repo, tag)
	},
}

func init() {
	appsCmd.AddCommand(appsAddCmd)
}

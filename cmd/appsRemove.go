package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

// appsRemoveCmd represents the appsRemove command
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

func init() {
	appsCmd.AddCommand(appsRemoveCmd)
}

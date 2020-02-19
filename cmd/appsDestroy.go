package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

// appsDestroyCmd represents the appsDestroy command
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

func init() {
	appsCmd.AddCommand(appsDestroyCmd)
}

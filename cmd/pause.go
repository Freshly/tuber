package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "pause [app name]",
	Short:        "pause deploys for the specified app",
	Args:         cobra.ExactArgs(1),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		return core.PauseDeployments(appName)
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}

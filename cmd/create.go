package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [app name] [docker repo] [deploy tag]",
	Short: "create new app in current cluster",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		repo := args[1]
		tag := args[2]

		return core.CreateTuberApp(appName, repo, tag)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

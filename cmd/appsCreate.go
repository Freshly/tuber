package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

// appsCreateCmd represents the appsCreate command
var appsCreateCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "create [app name] [docker repo] [deploy tag]",
	Short:        "create a new app in the current cluster",
	Args:         cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		repo := args[1]
		tag := args[2]

		return core.CreateTuberApp(appName, repo, tag)
	},
}

func init() {
	appsCmd.AddCommand(appsCreateCmd)
}

package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "create [docker repo] [deploy tag]",
	Short:        "create new app in current cluster",
	Args:         cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := args[1]
		tag := args[2]

		return core.CreateTuberApp(AppName, repo, tag)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

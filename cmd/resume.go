package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "resume [app name]",
	Short:        "resume deploys for the specified app",
	Args:         cobra.ExactArgs(1),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		return core.ResumeDeployments(appName)
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

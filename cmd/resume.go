package cmd

import (
	"tuber/pkg/k8s"

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

		exists, err := k8s.Exists("configmap", "tuber-app-pauses", "tuber")
		if err != nil {
			return err
		}

		if !exists {
			return nil
		}

		return k8s.RemoveConfigMapEntry("tuber-app-pauses", "tuber", appName)
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

package cmd

import (
	"tuber/pkg/k8s"

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

		exists, err := k8s.Exists("configmap", "tuber-app-pauses", "tuber")
		if err != nil {
			return err
		}

		if !exists {
			err = k8s.Create("tuber", "configmap", "tuber-app-pauses")
			if err != nil {
				return err
			}
		}

		return k8s.PatchConfigMap("tuber-app-pauses", "tuber", appName, "true")
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "auth",
	Short:        "authorize cli",
	Args:         cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}

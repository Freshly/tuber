package cmd

import (
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps [command]",
	Short: "A root command for app configurating.",
}

func init() {
	rootCmd.AddCommand(appsCmd)
}

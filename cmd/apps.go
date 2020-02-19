package cmd

import (
	"github.com/spf13/cobra"
)

// appsCmd represents the apps command
var appsCmd = &cobra.Command{
	Use:   "apps [command]",
	Short: "A root command for app configurating.",
}

func init() {
	rootCmd.AddCommand(appsCmd)
}

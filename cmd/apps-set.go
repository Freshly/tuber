package cmd

import (
	"github.com/spf13/cobra"
)

var appsSetCmd = &cobra.Command{
	Use:   "set [command]",
	Short: "a root command for setting top-level app fields.",
}

func init() {
	appsCmd.AddCommand(appsSetCmd)
}

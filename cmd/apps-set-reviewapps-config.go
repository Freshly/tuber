package cmd

import (
	"github.com/spf13/cobra"
)

var appsSetReviewappsConfigCmd = &cobra.Command{
	Use:   "rac [command]",
	Short: "a root command for setting ReviewAppsConfig fields (settings for creating new review apps)",
}

func init() {
	appsSetCmd.AddCommand(appsSetReviewappsConfigCmd)
}

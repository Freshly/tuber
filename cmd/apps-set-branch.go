package cmd

import (
	"github.com/freshly/tuber/pkg/gcr"
	"github.com/spf13/cobra"
)

var appsSetBranchCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "branch [app name] [branch]",
	Short:         "update the tag (branch) tuber will watch and deploy",
	Args:          cobra.ExactArgs(2),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetBranchCmd,
}

func runAppsSetBranchCmd(cmd *cobra.Command, args []string) error {
	appName := args[0]
	branch := args[1]
	app, err := getApp(appName)
	if err != nil {
		return err
	}

	imageTag, err := gcr.SwapTags(app.ImageTag, branch)
	if err != nil {
		return err
	}

	return runSetImageTagCmd(cmd, []string{appName, imageTag})
}

func init() {
	appsSetCmd.AddCommand(appsSetBranchCmd)
}

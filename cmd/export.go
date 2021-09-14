package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "export [app name]",
	Short:        "export editable and importable version of an app",
	Args:         cobra.ExactArgs(1),
	PreRunE:      displayCurrentContext,
	RunE:         runExportCmd,
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func runExportCmd(cmd *cobra.Command, args []string) error {
	appName := args[0]
	app, err := getApp(appName)
	if err != nil {
		return err
	}

	out, err := json.Marshal(app)
	if err != nil {
		return err
	}

	err = os.WriteFile(appName+".json", out, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

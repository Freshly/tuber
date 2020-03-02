package cmd

import (
	"fmt"
	"strings"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "exec [appName]",
	Short:        "execs a command on an app",
	RunE:         exec,
}

func exec(cmd *cobra.Command, args []string) error {
	jsonpath := fmt.Sprintf(`-o=jsonpath="%s"`, `{.items[0].metadata.name}`)
	name, err := k8s.GetCollection("pods", appName, jsonpath)
	if err != nil {
		return err
	}

	err = k8s.Exec(strings.Trim(string(name), "\""), appName, args...)

	return err
}

func init() {
	execCmd.PersistentFlags().StringVarP(&appName, "app", "a", "", "app name (required)")
	execCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(execCmd)
}

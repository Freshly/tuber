package cmd

import (
	"fmt"

	"github.com/freshly/tuber/pkg/config"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "context",
	Short:        "displays current context",
	RunE:         currentContext,
}

func currentContext(*cobra.Command, []string) error {
	config, err := config.Load()
	if err != nil {
		return err
	}

	cluster, err := config.CurrentClusterConfig()
	if err != nil {
		return err
	}

	fmt.Println(cluster.Shorthand)
	return nil
}

func init() {
	rootCmd.AddCommand(contextCmd)
}

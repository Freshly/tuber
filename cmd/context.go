package cmd

import (
	"fmt"

	"github.com/freshly/tuber/pkg/config"
	"github.com/freshly/tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "context",
	Short:        "displays current context",
	RunE:         currentContext,
}

func currentContext(*cobra.Command, []string) error {
	currentCluster, err := k8s.CurrentCluster()
	if err != nil {
		return err
	}

	config, err := config.Load()
	if err != nil {
		fmt.Println(currentCluster)
		return nil
	}

	if config == nil {
		fmt.Println(currentCluster)
		return nil
	}

	cluster := config.FindByName(currentCluster)

	if cluster.Name == "" {
		fmt.Println(currentCluster)
	} else {
		fmt.Println(cluster.Shorthand)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(contextCmd)
}

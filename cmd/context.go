package cmd

import (
	"fmt"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "context",
	Short:        "displays current context",
	RunE:         currentContext,
}

func currentContext(*cobra.Command, []string) error {
	config, err := getTuberConfig()
	if err != nil {
		return err
	}

	currentCluster, err := k8s.CurrentCluster()
	if err != nil {
		return err
	}

	if config == nil {
		return fmt.Errorf("tuber config empty, run `tuber config`")
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

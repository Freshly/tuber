package cmd

import (
	"fmt"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var switchClusterCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "switch [cluster name or shorthand]",
	Short:        "switch kubectl context to a different cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         switchCluster,
}

func switchCluster(cmd *cobra.Command, args []string) error {
	tuberrc, err := getTuberrc()
	if err != nil {
		return err
	}

	var clusterName string
	var displayCluster string

	clusterName = args[0]
	displayCluster = clusterName

	if tuberrc != nil {
		resolvedShorthand := tuberrc.Clusters[args[0]]
		if resolvedShorthand != "" {
			clusterName = resolvedShorthand
		}
	}

	err = k8s.UseCluster(clusterName)
	if err != nil {
		return err
	}

	fmt.Println("Switched to", displayCluster)

	return nil
}

func init() {
	rootCmd.AddCommand(switchClusterCmd)
}

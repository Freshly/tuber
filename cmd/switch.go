package cmd

import (
	"fmt"
	osExec "os/exec"

	"github.com/freshly/tuber/pkg/config"
	"github.com/freshly/tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var switchClusterCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "switch [shorthand cluster name]",
	Short:        "switch kubectl context to a different cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         switchCluster,
}

func switchCluster(cmd *cobra.Command, args []string) error {
	clusterShortName := args[0]

	config, err := config.Load()
	if err != nil {
		return err
	}

	cluster, err := config.FindByShortName(clusterShortName)
	if err != nil {
		return fmt.Errorf("cluster name not found")
	}

	currentCluster, err := config.CurrentClusterConfig()
	if err != nil {
		return err
	}

	if currentCluster.Shorthand == clusterShortName {
		fmt.Println("Already on", clusterShortName)
		return nil
	}

	k8sCheckErr := osExec.Command("kubectl", "version", "--client").Run()
	k8sPresent := k8sCheckErr == nil

	err = config.SetActive(cluster)
	if err != nil {
		return err
	}

	if k8sPresent {
		err = k8s.UseCluster(cluster.Name)
		if err != nil {
			return err
		}
	}

	fmt.Println("Switched to", clusterShortName)
	return nil
}

func init() {
	rootCmd.AddCommand(switchClusterCmd)
}

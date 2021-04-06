package cmd

import (
	"fmt"
	"strings"

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
	config, err := getTuberConfig()
	if err != nil {
		return err
	}

	currentCluster, err := k8s.CurrentCluster()
	if err != nil {
		return err
	}

	clusterShortName := args[0]

	if config == nil {
		return fmt.Errorf("tuber config empty, run `tuber config`")
	}

	cluster := config.FindByShortName(clusterShortName)

	if cluster.Name == "" {
		return fmt.Errorf("cluster name not found")
	}

	if strings.Trim(currentCluster, "\r\n") == cluster.Name {
		fmt.Println("Already on", clusterShortName)
	} else {
		err = k8s.UseCluster(cluster.Name)
		if err != nil {
			return err
		}
		fmt.Println("Switched to", clusterShortName)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(switchClusterCmd)
}

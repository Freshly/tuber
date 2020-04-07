package cmd

import (
	"fmt"
	"tuber/pkg/core"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var plantCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "plant [service account credentials path]",
	Short:        "tuberize a cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         plant,
}

func plant(cmd *cobra.Command, args []string) error {
	existsAlready, err := k8s.Exists("namespace", "tuber", "tuber")
	if err != nil {
		return err
	}

	if existsAlready {
		return fmt.Errorf("tuber already planted")
	}

	credentialsPath := args[0]
	err = core.NewAppSetup("tuber", false)
	if err != nil {
		return err
	}

	err = k8s.Create("tuber", "configmap", "tuber-apps")
	if err != nil {
		return err
	}

	return k8s.CreateTuberCredentials(credentialsPath, "tuber")
}

func init() {
	rootCmd.AddCommand(plantCmd)
}
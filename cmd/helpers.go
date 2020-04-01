package cmd

import (
	"fmt"
	"io/ioutil"
	"tuber/pkg/core"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func clusterData() (data *core.ClusterData) {
	return &core.ClusterData{
		DefaultGateway: viper.GetString("cluster-default-gateway"),
		DefaultHost:    viper.GetString("cluster-default-host"),
	}
}

func credentials() (creds []byte, err error) {
	viper.SetDefault("credentials-path", "/etc/tuber-credentials/credentials.json")
	credentialsPath := viper.GetString("credentials-path")
	creds, err = ioutil.ReadFile(credentialsPath)
	return
}

func promptCurrentContext(cmd *cobra.Command, args []string) error {
	skipConfirmation, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return err
	}

	if skipConfirmation {
		return nil
	}

	cluster, err := k8s.CurrentCluster()
	if err != nil {
		return err
	}
	fmt.Print("About to run ", cmd.Name(), " on ", cluster)
	fmt.Println("Press cmd+C / ctrl+C to cancel, enter to continue...")
	var input string
	_, err = fmt.Scanln(&input)

	if err != nil {
		if err.Error() == "unexpected newline" {
			return nil
		} else if err.Error() == "EOF" {
			return fmt.Errorf("cancelled")
		} else {
			return err
		}
	}
	return nil
}

func displayCurrentContext(cmd *cobra.Command, args []string) error {
	cluster, err := k8s.CurrentCluster()
	if err != nil {
		return err
	}
	fmt.Println("Running", cmd.Name(), "on", cluster)
	return nil
}

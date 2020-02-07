package cmd

import (
	"encoding/base64"
	"fmt"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use: "env [set || unset || file]",
}

var envSetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "set [appName] [key] [value]",
	RunE:         envSet,
	Args:         cobra.ExactArgs(3),
}

var envUnsetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "unset [appName] [key]",
	RunE:         envUnset,
	Args:         cobra.ExactArgs(2),
}

var fileCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "file [app] [local filepath]",
	Short:        "batch env set",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := k8s.CreateEnvFromFile(args[0], args[1])
		if err != nil {
			return err
		}
		return k8s.Restart("deployments", args[0])
	},
}

var envGetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "get [appName]",
	RunE:         envGet,
	Args:         cobra.ExactArgs(1),
}

func envSet(cmd *cobra.Command, args []string) error {
	appName := args[0]
	key := args[1]
	value := args[2]
	mapName := fmt.Sprintf("%s-env", appName)
	err := k8s.PatchSecret(mapName, appName, key, value)
	if err != nil {
		return err
	}
	return k8s.Restart("deployments", appName)
}

func envUnset(cmd *cobra.Command, args []string) error {
	appName := args[0]
	key := args[1]
	mapName := fmt.Sprintf("%s-env", appName)
	err := k8s.RemoveSecretEntry(mapName, appName, key)
	if err != nil {
		return err
	}
	return k8s.Restart("deployments", appName)
}

func envGet(cmd *cobra.Command, args []string) (err error) {
	appName := args[0]
	mapName := fmt.Sprintf("%s-env", appName)
	config, err := k8s.GetConfig(mapName, appName, "Secret")
	if err != nil {
		return
	}
	for k, v := range config.Data {
		decoded, decodeErr := base64.StdEncoding.DecodeString(v)
		if decodeErr != nil {
			return
		}
		fmt.Println(k+":", string(decoded))
	}
	return
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(fileCmd)
	envCmd.AddCommand(envGetCmd)
}

package cmd

import (
	"encoding/base64"
	"fmt"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var envCmd = &cobra.Command{
	Use: "env [set || unset || file]",
}

var envSetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "set [key] [value]",
	RunE:         envSet,
	Args:         cobra.ExactArgs(2),
}

var envUnsetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "unset [key]",
	RunE:         envUnset,
	Args:         cobra.ExactArgs(1),
}

var envGetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "get [key]",
	Args:         cobra.ExactArgs(1),
	RunE:         envGet,
}

var fileCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "file [local filepath]",
	Short:        "batch env set",
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := k8s.CreateEnvFromFile(appName, args[1])
		if err != nil {
			return err
		}
		return k8s.Restart("deployments", appName)
	},
}

var envListCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "list",
	RunE:         envList,
}

func envSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]
	mapName := fmt.Sprintf("%s-env", appName)

	logger, err := createLogger()
	if err != nil {
		return err
	}

	logger.Info("env: set",
		zap.String("name", appName),
		zap.String("key", key),
		zap.String("action", "change_env"),
	)

	err = k8s.PatchSecret(mapName, appName, key, value)
	if err != nil {
		return err
	}
	return k8s.Restart("deployments", appName)
}

func envUnset(cmd *cobra.Command, args []string) error {
	key := args[0]
	mapName := fmt.Sprintf("%s-env", appName)

	logger, err := createLogger()
	if err != nil {
		return err
	}

	logger.Info("env: unset",
		zap.String("name", appName),
		zap.String("key", key),
		zap.String("action", "change_env"),
	)

	err = k8s.RemoveSecretEntry(mapName, appName, key)
	if err != nil {
		return err
	}

	return k8s.Restart("deployments", appName)
}

func envGet(cmd *cobra.Command, args []string) (err error) {
	mapName := fmt.Sprintf("%s-env", appName)
	key := args[0]
	config, err := k8s.GetConfig(mapName, appName, "Secret")

	if err != nil {
		return
	}

	v := config.Data[key]
	decoded, err := base64.StdEncoding.DecodeString(v)

	if err != nil {
		return
	}

	fmt.Println(string(decoded))
	return
}

func envList(cmd *cobra.Command, _ []string) (err error) {
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
	envCmd.PersistentFlags().StringVarP(&appName, "app", "a", "", "app name (required)")
	envCmd.MarkPersistentFlagRequired("app")

	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(fileCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envListCmd)
}

package cmd

import (
	"encoding/base64"
	"fmt"
	"sort"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var envCmd = &cobra.Command{
	Use: "env [set || unset || get || list || file]",
}

var envSetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "set [app] [key] [value]",
	RunE:         envSet,
	Args:         cobra.ExactArgs(3),
	PreRunE:      promptCurrentContext,
}

var envUnsetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "unset [app] [key]",
	RunE:         envUnset,
	Args:         cobra.ExactArgs(2),
	PreRunE:      promptCurrentContext,
}

var envGetCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "get [app] [key]",
	Args:         cobra.ExactArgs(2),
	RunE:         envGet,
	PreRunE:      displayCurrentContext,
}

var fileCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "file [app] [local filepath]",
	Short:        "batch env set",
	Args:         cobra.ExactArgs(2),
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		err := k8s.CreateEnvFromFile(appName, args[1])
		if err != nil {
			return err
		}
		return k8s.Restart("deployments", appName)
	},
}

var envListCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "list [app]",
	Short:        "decode and display an app's env",
	RunE:         envList,
	Args:         cobra.ExactArgs(1),
	PreRunE:      displayCurrentContext,
}

func envSet(cmd *cobra.Command, args []string) error {
	appName := args[0]
	key := args[1]
	value := args[2]
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
	appName := args[0]
	key := args[1]
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
	appName := args[0]
	mapName := fmt.Sprintf("%s-env", appName)
	key := args[1]
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

func envList(cmd *cobra.Command, args []string) error {
	appName := args[0]
	mapName := fmt.Sprintf("%s-env", appName)
	config, err := k8s.GetConfig(mapName, appName, "Secret")
	if err != nil {
		return err
	}

	var list []string
	for k, v := range config.Data {
		decoded, decodeErr := base64.StdEncoding.DecodeString(v)
		if decodeErr != nil {
			return decodeErr
		}
		list = append(list, k+`: "`+string(decoded)+`"`)
	}

	sort.Strings(list)
	for _, v := range list {
		fmt.Println(v)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
	envCmd.AddCommand(fileCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envListCmd)
}

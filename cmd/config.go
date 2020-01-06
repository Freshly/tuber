package cmd

import (
	"fmt"
	"log"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use: "config [patch || remove]",
}

var configAddCmd = &cobra.Command{
	Use:  "add [appName] [key] [value]",
	Run:  configAdd,
	Args: cobra.ExactArgs(3),
}

var configRemoveCmd = &cobra.Command{
	Use:  "remove [appName] [key]",
	Run:  configRemove,
	Args: cobra.ExactArgs(2),
}

func configAdd(cmd *cobra.Command, args []string) {
	appName := args[0]
	key := args[1]
	value := args[2]
	mapName := fmt.Sprintf("%s-config", appName)
	err := k8s.PatchConfig(mapName, appName, key, value)
	if err != nil {
		log.Fatal(err)
	}
}

func configRemove(cmd *cobra.Command, args []string) {
	appName := args[0]
	key := args[1]
	mapName := fmt.Sprintf("%s-config", appName)
	err := k8s.RemoveConfigEntry(mapName, appName, key)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
}

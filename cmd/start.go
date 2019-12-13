package cmd

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"tuber/pkg/listen"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tuber",
	Run:   start,
}

func start(cmd *cobra.Command, args []string) {
	godotenv.Load()

	listen.Listen(func(event *listen.RegistryEvent, err error) {
		spew.Dump(event)
	})
}

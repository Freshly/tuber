package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the client version information",
	RunE:  info,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func info(cmd *cobra.Command, args []string) error {
	fmt.Println(Version)

	return nil
}

package cmd

import (
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var credentialsCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "credentials [local filepath] [namespace]",
	Short:        "add tuber secrets from file",
	PreRunE:      promptCurrentContext,
	RunE: func(cmd *cobra.Command, args []string) error {
		return k8s.CreateTuberCredentials(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(credentialsCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var documentCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "document",
	Short:        "Generates documentation for tuber",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := doc.GenMarkdownTree(rootCmd, "./doc")
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(documentCmd)
}

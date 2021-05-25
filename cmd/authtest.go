package cmd

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"

	"github.com/spf13/cobra"
)

// authtestCmd represents the authtest command
var authtestCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "authtest",
	Short:        "display access token",
	Long:         `print an access token for debugging purposes`,
	RunE:         authtest,
}

func init() {
	rootCmd.AddCommand(authtestCmd)
}

func authtest(cmd *cobra.Command, args []string) (err error) {
	creds, err := google.FindDefaultCredentials(context.Background())
	if err != nil {
		return
	}

	fmt.Println(string(creds.JSON))

	return
}

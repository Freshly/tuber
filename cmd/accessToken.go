package cmd

import (
	"fmt"

	"github.com/freshly/tuber/pkg/gcloud"

	"github.com/spf13/cobra"
)

// accessTokenCmd represents the accessToken command
var accessTokenCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "access-token",
	Short:        "display access token",
	Long:         `print an access token for debugging purposes`,
	RunE:         accessToken,
}

func init() {
	rootCmd.AddCommand(accessTokenCmd)
}

func accessToken(cmd *cobra.Command, args []string) (err error) {
	fmt.Println("accessToken called")
	creds, err := credentials()
	if err != nil {
		return
	}

	token, err := gcloud.GetAccessToken(creds)

	if err != nil {
		return
	}

	fmt.Printf("Token is:\n%s\n", token)

	return
}

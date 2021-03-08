package cmd

import (
	"tuber/pkg/adminserver"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var adminserverCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "adminserver",
	Short:        "starts the admin http server for review apps and maybe other stuff who knows",
	Run:          startAdminServer,
}

func startAdminServer(cmd *cobra.Command, args []string) {
	logger, err := createLogger()
	defer logger.Sync()

	if err != nil {
		panic(err)
	}

	projectName := viper.GetString("review-apps-triggers-project-name")
	if projectName == "" {
		panic("need a review apps triggers project name")
	}

	creds, err := credentials()
	if err != nil {
		panic(err)
	}
	adminserver.Start(projectName, creds, logger)
}

func init() {
	rootCmd.AddCommand(adminserverCmd)
}

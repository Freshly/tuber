package cmd

import (
	"context"
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
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	logger, err := createLogger()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	triggersProjectName := viper.GetString("review-apps-triggers-project-name")
	if triggersProjectName == "" {
		panic("need a review apps triggers project name")
	}

	creds, err := credentials()
	if err != nil {
		panic(err)
	}
	err = adminserver.Start(ctx, logger, triggersProjectName, creds, viper.GetBool("reviewapps-enabled"), viper.GetString("cluster-default-host"))
	if err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(adminserverCmd)
}

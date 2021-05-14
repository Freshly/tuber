package cmd

import (
	"context"

	"github.com/freshly/tuber/pkg/adminserver"

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

	db, err := db()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	logger, err := createLogger()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	reviewAppsEnabled := viper.GetBool("reviewapps-enabled")

	triggersProjectName := viper.GetString("review-apps-triggers-project-name")
	if reviewAppsEnabled && triggersProjectName == "" {
		panic("need a review apps triggers project name")
	}

	creds, err := credentials()
	if err != nil {
		panic(err)
	}
	err = adminserver.Start(ctx, logger, db, triggersProjectName, creds, viper.GetBool("reviewapps-enabled"), viper.GetString("cluster-default-host"), viper.GetString("adminserver-port"))
	if err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(adminserverCmd)
}

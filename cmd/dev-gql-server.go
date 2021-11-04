package cmd

import (
	"context"

	"github.com/freshly/tuber/pkg/config"
	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var devGqlServerCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "dev-gql-server [appname]",
	Short:         "for local development only! start a gql server with a tuber app",
	RunE:          runDevGqlServer,
	PreRunE:       promptCurrentContext,
	Args:          cobra.ExactArgs(1),
}

func runDevGqlServer(cmd *cobra.Command, args []string) error {
	app, err := getApp(args[0])
	if err != nil {
		return err
	}
	db, err := openDB()
	if err != nil {
		return err
	}
	err = db.SaveApp(app)
	if err != nil {
		return err
	}

	creds, err := credentials()
	if err != nil {
		return err
	}

	logger, err := createLogger()
	if err != nil {
		return err
	}
	defer logger.Sync()

	ctx := context.Background()

	data, err := clusterData()
	if err != nil {
		return err
	}

	config, err := config.Load()
	if err != nil {
		return err
	}
	cc, err := config.CurrentClusterConfig()
	if err != nil {
		return err
	}

	slackClient := slack.New("", false, "")

	viper.SetDefault("TUBER_USE_DEVSERVER", true)
	viper.SetDefault("TUBER_ADMINSERVER_PORT", "3000")
	b := make([]rune, 32)
	for i := range b {
		b[i] = []rune("a")[0]
	}
	viper.SetDefault("TUBER_COOKIE_BLOCK_KEY", string(b))
	viper.SetDefault("TUBER_COOKIE_HASH_KEY", "asdfasdf")
	viper.SetDefault("TUBER_CLUSTER_REGION", "us-central1")
	viper.SetDefault("TUBER_CLUSTER_NAME", cc.Shorthand)
	viper.SetDefault("TUBER_DEBUG", true)
	viper.SetDefault("TUBER_REVIEWAPPS_ENABLED", true)
	processor := events.NewProcessor(ctx, logger, db, creds, data, true, slackClient, "", "", "")
	startAdminServer(ctx, db, processor, logger, creds)

	return nil
}

func init() {
	rootCmd.AddCommand(devGqlServerCmd)
}

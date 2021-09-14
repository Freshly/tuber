package cmd

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/builds"
	"github.com/freshly/tuber/pkg/pubsub"
	"github.com/freshly/tuber/pkg/report"
	"github.com/freshly/tuber/pkg/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(startBuildsCmd)
}

var startBuildsCmd = &cobra.Command{
	Use:   "startBuilds",
	Short: "start the builds pubsub",
	RunE:  startBuilds,
}

func startBuilds(cmd *cobra.Command, args []string) error {
	logger, err := createLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	scope := report.Scope{"during": "startup"}
	startupLogger := logger.With(zap.String("action", "startup"))

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	creds, err := credentials()
	if err != nil {
		startupLogger.Warn("failed to get credentials", zap.Error(err))
		report.Error(err, scope.WithContext("getting credentials"))
		panic(err)
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, appName := range []string{"payment-service", "catalog-service", "potatoes"} {
		var app *model.TuberApp
		app, err = getApp(appName)
		if err != nil {
			return err
		}

		app.SlackChannel = "lauren-sneaky-channel"

		err = db.SaveApp(app)
		if err != nil {
			return err
		}
	}

	data, err := clusterData()
	if err != nil {
		startupLogger.Warn("failed to get cluster data", zap.Error(err))
		report.Error(err, scope.WithContext("getting cluster data"))
		panic(err)
	}

	fmt.Println(viper.GetString("TUBER_PUBSUB_PROJECT"))
	fmt.Println(viper.GetString("TUBER_PUBSUB_CLOUDBUILD_SUBSCRIPTION_NAME"))

	slackClient := slack.New(viper.GetString("TUBER_SLACK_TOKEN"), viper.GetBool("TUBER_SLACK_ENABLED"), viper.GetString("TUBER_SLACK_CATCHALL_CHANNEL"))

	buildEventProcessor := builds.NewProcessor(ctx, logger, db, slackClient)
	buildListener, err := pubsub.NewListener(
		ctx,
		logger,
		viper.GetString("TUBER_PUBSUB_PROJECT"),
		viper.GetString("TUBER_PUBSUB_CLOUDBUILD_SUBSCRIPTION_NAME"),
		creds,
		data,
		buildEventProcessor,
	)
	if err != nil {
		startupLogger.Error("failed to start cloud build listener", zap.Error(err))
		report.Error(err, scope.WithContext("initialize cloud build listener"))
		panic(err)
	}

	err = buildListener.Start()
	if err != nil {
		panic(err)
	}

	return nil
}

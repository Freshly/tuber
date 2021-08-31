package cmd

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/pkg/builds"
	"github.com/freshly/tuber/pkg/pubsub"
	"github.com/freshly/tuber/pkg/report"
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

	data, err := clusterData()
	if err != nil {
		startupLogger.Warn("failed to get cluster data", zap.Error(err))
		report.Error(err, scope.WithContext("getting cluster data"))
		panic(err)
	}

	fmt.Println(viper.GetString("TUBER_PUBSUB_PROJECT"))
	fmt.Println(viper.GetString("TUBER_PUBSUB_CLOUDBUILD_SUBSCRIPTION_NAME"))

	buildEventProcessor := builds.NewProcessor()
	buildListener, err := pubsub.NewListener(
		ctx,
		logger,
		viper.GetString("TUBER_PUBSUB_PROJECT"),
		"tuber-test-sub",
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

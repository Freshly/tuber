package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/pubsub"
	"github.com/freshly/tuber/pkg/report"
	"github.com/freshly/tuber/pkg/slack"
	"github.com/getsentry/sentry-go"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tuber's pub/sub server",
	RunE:  start,
}

// Attaches interrupt and terminate signals to a cancel function
func bindShutdown(logger *zap.Logger, cancel func()) {
	var signals = make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		s := <-signals
		logger.With(zap.Reflect("signal", s)).Info("Signal received")
		cancel()
	}()
}

func start(cmd *cobra.Command, args []string) error {
	logger, err := createLogger()
	defer logger.Sync()

	if err != nil {
		return err
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	initErrorReporters()
	defer sentry.Recover()

	scope := report.Scope{"during": "startup"}
	startupLogger := logger.With(zap.String("action", "startup"))

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	bindShutdown(logger, cancel)

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

	slackClient := slack.New(viper.GetString("slack-token"), viper.GetBool("slack-enabled"), viper.GetString("slack-catchall-channel"))
	processor := events.NewProcessor(ctx, logger, db, creds, data, viper.GetBool("reviewapps-enabled"), slackClient, viper.GetString("sentry-bearer-token"))
	listener, err := pubsub.NewListener(
		ctx,
		logger,
		viper.GetString("pubsub-project"),
		viper.GetString("pubsub-subscription-name"),
		creds,
		data,
		processor,
	)

	if err != nil {
		startupLogger.Warn("failed to initialize listener", zap.Error(err))
		report.Error(err, scope.WithContext("initialize listener"))
		panic(err)
	}

	go startAdminServer(ctx, db, processor, logger, creds)

	err = listener.Start()
	if err != nil {
		startupLogger.Warn("listener shutdown", zap.Error(err))
		report.Error(err, scope.WithContext("listener shutdown"))
		panic(err)
	}

	<-ctx.Done()
	logger.Info("Shutting down...")
	return nil
}

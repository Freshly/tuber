package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"tuber/pkg/events"
	"tuber/pkg/listener"
	"tuber/pkg/reviewapps"
	"tuber/pkg/sentry"
	"tuber/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start tuber's pub/sub listener",
	Run:     start,
	PreRunE: promptCurrentContext,
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

func start(cmd *cobra.Command, args []string) {
	// Create a logger and defer an final sync (os.flush())
	logger, err := createLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Report any errors to Sentry
	sentryEnabled := viper.GetBool("sentry-enabled")
	sentryDsn := viper.GetString("sentry-dsn")
	errReports := make(chan error, 1)

	defer close(errReports)

	go sentry.Stream(sentryEnabled, sentryDsn, errReports, logger)

	// calling cancel() will signal to the rest of the application
	// that we want to shut down
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// bind the cancel to signals
	bindShutdown(logger, cancel)

	// create a new PubSub listener
	var options = make([]listener.Option, 0)
	if viper.IsSet("max-accept") {
		options = append(options, listener.WithMaxAccept(viper.GetInt("max-accept")))
	}

	if viper.IsSet("max-timeout") {
		options = append(options, listener.WithMaxTimeout(viper.GetDuration("max-timeout")))
	}

	subscriptionName := viper.GetString("pubsub-subscription-name")
	if subscriptionName == "" {
		panic(errors.New("pubsub subscription name is required"))
	}

	var l = listener.NewListener(logger, subscriptionName, options...)

	creds, err := credentials()
	if err != nil {
		panic(err)
	}

	unprocessedEvents, processedEvents, failedEvents, err := l.Listen(ctx, creds)
	if err != nil {
		panic(err)
	}

	data, err := clusterData()
	if err != nil {
		panic(err)
	}

	eventProcessor := events.EventProcessor{
		Creds:             creds,
		Logger:            logger,
		ClusterData:       data,
		ReviewAppsEnabled: viper.GetBool("reviewapps-enabled"),
		Unprocessed:       unprocessedEvents,
		Processed:         processedEvents,
		ChErr:             failedEvents,
		ChErrReports:      errReports,
	}
	go eventProcessor.Start()

	go func() {
		logger = logger.With(zap.String("action", "grpc"))

		srv := reviewapps.Server{
			ReviewAppsEnabled:  viper.GetBool("reviewapps-enabled"),
			ClusterDefaultHost: viper.GetString("cluster-default-host"),
			ProjectName:        viper.GetString("project-name"),
			Logger:             logger,
			Credentials:        creds,
		}

		err = server.Start(3000, srv)
		if err != nil {
			logger.Error("grpc server: failed to start")
			cancel()
		}
	}()

	// Wait for cancel() of context
	<-ctx.Done()
	logger.Info("Shutting down...")

	// Wait for queues to drain
	l.Wait()
}

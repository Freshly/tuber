package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"tuber/pkg/events"
	"tuber/pkg/listener"

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
	logger, err := createLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	errReports := errorReportingChannel(logger)

	defer close(errReports)

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// bind the cancel to signals
	bindShutdown(logger, cancel)

	subscriptionName := viper.GetString("pubsub-subscription-name")
	if subscriptionName == "" {
		panic(errors.New("pubsub subscription name is required"))
	}

	var l = listener.NewListener(logger, subscriptionName)

	creds, err := credentials()
	if err != nil {
		panic(err)
	}

	unprocessedEvents, processedEvents, failedReleases, err := l.Listen(ctx, creds)
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
		FailedReleases:    failedReleases,
		Errors:            errReports,
	}
	go eventProcessor.Start()

	// Wait for cancel() of context
	<-ctx.Done()
	logger.Info("Shutting down...")

	// Wait for queues to drain
	l.Wait()
}

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"tuber/pkg/errors"
	"tuber/pkg/events"
	"tuber/pkg/gcloud"
	"tuber/pkg/listener"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tuber",
	Run:   start,
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
	logger := createLogger()
	defer logger.Sync()

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

	var l = listener.NewListener(logger, options...)

	// Report any errors to Sentry
	sentryEnabled := viper.GetBool("sentry-enabled")
	sentryDsn := viper.GetString("sentry-dsn")
	errReports := make(chan error, 1)

	defer close(errReports)

	if sentryEnabled {
		errors.InitSentry(sentryDsn, errReports)
	}

	creds, err := credentials()
	if err != nil {
		panic(err)
	}

	unprocessedEvents, processedEvents, failedEvents, err := l.Listen(ctx, creds)
	if err != nil {
		panic(err)
	}

	token, err := gcloud.GetAccessToken(creds)

	if err != nil {
		panic(err)
	}

	// Create a new streamer
	streamer := events.NewStreamer(token, logger)
	go streamer.Stream(unprocessedEvents, processedEvents, failedEvents, errReports)

	// Wait for cancel() of context
	<-ctx.Done()
	logger.Info("Shutting down...")

	// Wait for queues to drain
	l.Wait()
}

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"tuber/pkg/k8s"

	"tuber/pkg/events"
	"tuber/pkg/listener"
	"tuber/pkg/sentry"

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

	subscriptionName := viper.GetString("pubsub-subscription-name")

	config, err := k8s.GetSecret("tuber", "tuber-env")
	if err != nil {
		panic(err)
	}

	if subscriptionName == config.Data["TUBER_PUBSUB_SUBSCRIPTION_NAME"] {
		panic(fmt.Errorf("do not run start on your cluster's subscription"))
	}

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

	if len(subscriptionName) < 1 {
		err = errors.New("pubsub subscription name is required")
		panic(err)
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

	// Create a new streamer
	streamer := events.NewStreamer(creds, logger, data)
	go streamer.Stream(unprocessedEvents, processedEvents, failedEvents, errReports)

	// Wait for cancel() of context
	<-ctx.Done()
	logger.Info("Shutting down...")

	// Wait for queues to drain
	l.Wait()
}

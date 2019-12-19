package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"tuber/pkg/events"
	"tuber/pkg/listener"
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

// Creates a channel that logs errors
func createErrorChannel(logger *zap.Logger) chan<- error {
	var errorChan = make(chan error, 1)
	go func() {
		logger.Info("error listener: started")
		for err := range errorChan {
			logger.Warn("error while processing", zap.Error(err))
		}
		logger.Info("error listener: shutdown")
	}()
	return errorChan
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

	unprocessedEvents, processedEvents, err := l.Listen(ctx)
	if err != nil {
		panic(err)
	}

	// Create error channel
	var errorChan = createErrorChannel(logger)

	var token = viper.GetString("gcloud-token")

	// Create a new streamer
	streamer := events.NewStreamer(token, logger)
	go streamer.Stream(unprocessedEvents, processedEvents, errorChan)

	// Wait for cancel() of context
	<-ctx.Done()
	logger.Info("Shutting down...")

	// Wait for queues to drain
	l.Wait()
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"tuber/pkg/events"
	"tuber/pkg/pubsub"
	"tuber/pkg/report"
	"tuber/pkg/reviewapps"
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
	Short:   "Start tuber's pub/sub server",
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

	initErrorReporters()
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

	listener, err := pubsub.NewListener(
		ctx,
		logger,
		viper.GetString("pubsub-project"),
		viper.GetString("pubsub-subscription-name"),
		creds,
		data,
		events.NewProcessor(ctx, logger, creds, data, viper.GetBool("reviewapps-enabled")),
	)

	if err != nil {
		startupLogger.Warn("failed to initialize listener", zap.Error(err))
		report.Error(err, scope.WithContext("initialize listener"))
		panic(err)
	}

	if viper.GetBool("reviewapps-enabled") {
		go startReviewAppsServer(logger, creds)
	}

	err = listener.Start()
	if err != nil {
		startupLogger.Warn("listener shutdown", zap.Error(err))
		report.Error(err, scope.WithContext("listener shutdown"))
		panic(err)
	}

	<-ctx.Done()
	logger.Info("Shutting down...")
}

func startReviewAppsServer(logger *zap.Logger, creds []byte) {
	logger = logger.With(zap.String("action", "grpc"))

	projectName := viper.GetString("review-apps-triggers-project-name")
	if projectName == "" {
		err := fmt.Errorf("no reviewapps project name configured")
		logger.Error("grpc server failed to start: no reviewapps project name configured")
		report.Error(err, report.Scope{"during": "grpc server startup"})
		panic(err)
	}

	srv := reviewapps.Server{
		ClusterDefaultHost: viper.GetString("cluster-default-host"),
		ProjectName:        projectName,
		Logger:             logger,
		Credentials:        creds,
	}

	logger.Debug("starting GRPC server")
	err := server.Start(3000, srv)
	if err != nil {
		logger.Error("grpc server failed to start")
		report.Error(err, report.Scope{"during": "grpc server startup"})
		panic(err)
	}
}

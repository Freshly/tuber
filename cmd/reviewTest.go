package cmd

import (
	"context"
	"tuber/pkg/reviewapps"
	"tuber/pkg/server"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var reviewTest = &cobra.Command{
	SilenceUsage: true,
	Use:          "review-test",
	Run: func(cmd *cobra.Command, args []string) {
		var _, cancel = context.WithCancel(context.Background())
		defer cancel()

		logger, err := createLogger()
		if err != nil {
			panic(err)
		}
		defer logger.Sync()

		creds, err := credentials()
		if err != nil {
			panic(err)
		}

		logger = logger.With(zap.String("action", "grpc"))

		srv := reviewapps.Server{
			ReviewAppsEnabled:  true,
			ClusterDefaultHost: "staging.freshlyservices.net",
			ProjectName:        "freshly-docker",
			Logger:             logger,
			Credentials:        creds,
		}

		err = server.Start(3000, srv)
		if err != nil {
			logger.Error("grpc server: failed to start")
			cancel()
		}
	},
}

func init() {
	rootCmd.AddCommand(reviewTest)
}

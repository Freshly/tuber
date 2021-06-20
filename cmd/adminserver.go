package cmd

import (
	"context"

	"github.com/freshly/tuber/pkg/adminserver"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/events"
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

func startAdminServer(ctx context.Context, db *core.DB, processor *events.Processor, logger *zap.Logger, creds []byte) {
	reviewAppsEnabled := viper.GetBool("reviewapps-enabled")

	triggersProjectName := viper.GetString("review-apps-triggers-project-name")
	if reviewAppsEnabled && triggersProjectName == "" {
		panic("need a review apps triggers project name")
	}

	err := adminserver.Start(ctx, logger, db, processor, triggersProjectName, creds,
		viper.GetBool("reviewapps-enabled"),
		viper.GetString("cluster-default-host"),
		viper.GetString("adminserver-port"),
		viper.GetString("cluster-name"),
		viper.GetString("cluster-region"),
	)

	if err != nil {
		panic(err)
	}
}

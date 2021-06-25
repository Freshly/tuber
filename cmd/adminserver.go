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
	reviewAppsEnabled := viper.GetBool("TUBER_REVIEWAPPS_ENABLED")

	triggersProjectName := viper.GetString("TUBER_REVIEW_APPS_TRIGGERS_PROJECT_NAME")
	if reviewAppsEnabled && triggersProjectName == "" {
		panic("need a review apps triggers project name")
	}

	viper.SetDefault("TUBER_ADMINSERVER_PREFIX", "/tuber")
	err := adminserver.Start(ctx, logger, db, processor, triggersProjectName, creds,
		reviewAppsEnabled,
		viper.GetString("TUBER_CLUSTER_DEFAULT_HOST"),
		viper.GetString("TUBER_ADMINSERVER_PORT"),
		viper.GetString("TUBER_CLUSTER_NAME"),
		viper.GetString("TUBER_CLUSTER_REGION"),
		viper.GetString("TUBER_ADMINSERVER_PREFIX"),
		viper.GetBool("TUBER_USE_DEVSERVER"),
	)

	if err != nil {
		panic(err)
	}
}

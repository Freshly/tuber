package cmd

import (
	"context"

	"github.com/freshly/tuber/pkg/adminserver"
	"github.com/freshly/tuber/pkg/core"
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

func startAdminServer(ctx context.Context, db *core.DB, logger *zap.Logger, creds []byte) {
	reviewAppsEnabled := viper.GetBool("reviewapps-enabled")

	triggersProjectName := viper.GetString("review-apps-triggers-project-name")
	if reviewAppsEnabled && triggersProjectName == "" {
		panic("need a review apps triggers project name")
	}

	if err := adminserver.Start(ctx, logger, db, triggersProjectName, creds, viper.GetBool("reviewapps-enabled"), viper.GetString("cluster-default-host"), viper.GetString("adminserver-port")); err != nil {
		panic(err)
	}
}

package cmd

import (
	"context"

	"github.com/freshly/tuber/pkg/adminserver"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/oauth"
	"github.com/gorilla/securecookie"
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

func startAdminServer(ctx context.Context, db *core.DB, processor *events.Processor, logger *zap.Logger, creds []byte) {
	reviewAppsEnabled := viper.GetBool("TUBER_REVIEWAPPS_ENABLED")

	triggersProjectName := viper.GetString("TUBER_REVIEW_APPS_TRIGGERS_PROJECT_NAME")
	if reviewAppsEnabled && triggersProjectName == "" {
		panic("need a review apps triggers project name")
	}

	auth := oauth.NewAuthenticator(viper.GetString("TUBER_OAUTH_REDIRECT_URL"), viper.GetString("TUBER_OAUTH_WEB_CLIENT_SECRET"), viper.GetString("TUBER_OAUTH_WEB_CLIENT_ID"), viper.GetString("TUBER_OAUTH_STATE_KEY"))
	if len(viper.GetString("TUBER_COOKIE_BLOCK_KEY")) < 32 {
		logger.Warn("starting admin server with TUBER_COOKIE_BLOCK_KEY set to a value under 32 characters. Use a 32 character value for aes-256.")
	}
	secureCookie := securecookie.New([]byte(viper.GetString("TUBER_COOKIE_HASH_KEY")), []byte(viper.GetString("TUBER_COOKIE_BLOCK_KEY")))

	viper.SetDefault("TUBER_ADMINSERVER_PREFIX", "/tuber")
	err := adminserver.Start(ctx, logger, db, processor, triggersProjectName, creds,
		reviewAppsEnabled,
		viper.GetString("TUBER_CLUSTER_DEFAULT_HOST"),
		viper.GetString("TUBER_ADMINSERVER_PORT"),
		viper.GetString("TUBER_CLUSTER_NAME"),
		viper.GetString("TUBER_CLUSTER_REGION"),
		viper.GetString("TUBER_ADMINSERVER_PREFIX"),
		viper.GetBool("TUBER_USE_DEVSERVER"),
		auth,
		secureCookie,
	)

	if err != nil {
		panic(err)
	}
}

package cmd

import (
	"context"

	"github.com/freshly/tuber/pkg/containers"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/slack"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "deploy [app]",
	Short:        "deploys the latest built image of an app",
	RunE:         deploy,
	PreRunE:      promptCurrentContext,
}

func deploy(cmd *cobra.Command, args []string) error {
	appName := args[0]
	logger, err := createLogger()
	if err != nil {
		return err
	}

	defer logger.Sync()

	apps, err := core.TuberSourceApps()

	if err != nil {
		return err
	}

	creds, err := credentials()
	if err != nil {
		return err
	}

	app, err := apps.FindApp(appName)
	if err != nil {
		return err
	}

	location := core.GetRepositoryLocation(app)

	sha, err := containers.GetLatestSHA(location, creds)

	if err != nil {
		return err
	}

	data, err := clusterData()
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	slackClient := slack.New(viper.GetString("slack-token"), viper.GetBool("slack-enabled"), viper.GetString("slack-catchall-channel"))
	processor := events.NewProcessor(ctx, logger, creds, data, viper.GetBool("reviewapps-enabled"), slackClient)
	digest := app.RepoHost + "/" + app.RepoPath + "@" + sha
	tag := app.ImageTag

	processor.ProcessMessage(digest, tag)
	return nil
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

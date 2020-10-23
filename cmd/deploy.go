package cmd

import (
	"context"
	"tuber/pkg/containers"
	"tuber/pkg/core"
	"tuber/pkg/events"
	"tuber/pkg/pubsub"

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

	location := app.GetRepositoryLocation()

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

	processor := events.NewProcessor(ctx, logger, creds, data, viper.GetBool("reviewapps-enabled"))

	digest := app.RepoHost + "/" + app.RepoPath + "@" + sha
	tag := app.ImageTag
	var message = new(pubsub.Message)
	message.Data = []byte(`{"action":"INSERT","digest":"` + digest + `","tag":"` + tag + `"}`)

	processor.ProcessMessage(message)
	return nil
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

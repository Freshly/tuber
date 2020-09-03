package cmd

import (
	"fmt"
	"tuber/pkg/containers"
	"tuber/pkg/core"
	"tuber/pkg/events"
	"tuber/pkg/listener"

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

type emptyAckable struct{}

func (emptyAckable) Ack()  {}
func (emptyAckable) Nack() {}

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

	unprocessed := make(chan *listener.RegistryEvent, 1)
	processed := make(chan *listener.RegistryEvent, 1)
	failedReleases := make(chan listener.FailedRelease, 1)
	errors := make(chan error, 1)

	eventProcessor := events.EventProcessor{
		Creds:             creds,
		Logger:            logger,
		ClusterData:       data,
		ReviewAppsEnabled: viper.GetBool("reviewapps-enabled"),
		Unprocessed:       unprocessed,
		Processed:         processed,
		FailedReleases:    failedReleases,
		Errors:            errors,
	}

	go eventProcessor.Start()

	ackable := emptyAckable{}
	deployEvent := listener.RegistryEvent{
		Action:  "INSERT",
		Digest:  app.RepoHost + "/" + app.RepoPath + "@" + sha,
		Tag:     app.ImageTag,
		Message: ackable,
	}

	unprocessed <- &deployEvent

	select {
	case <-failedReleases:
		close(unprocessed)
		return fmt.Errorf("deploy failed")
	case <-errors:
		close(unprocessed)
		return nil
	}
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

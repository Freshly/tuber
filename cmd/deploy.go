package cmd

import (
	"fmt"
	"tuber/pkg/containers"
	"tuber/pkg/core"
	"tuber/pkg/events"
	"tuber/pkg/listener"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "deploy",
	Short:        "Deploys an app",
	RunE:         deploy,
}

type emptyAckable struct{}

func (emptyAckable) Ack()  {}
func (emptyAckable) Nack() {}

func deploy(cmd *cobra.Command, args []string) error {
	logger, err := createLogger()
	if err != nil {
		return err
	}

	defer logger.Sync()

	apps, err := core.TuberApps()

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

	streamer := events.NewStreamer(creds, logger, clusterData())

	errorChan := make(chan listener.FailedRelease, 1)
	unprocessedEvents := make(chan *listener.RegistryEvent, 1)
	processedEvents := make(chan *listener.RegistryEvent, 1)
	errorReports := make(chan error, 1)

	go streamer.Stream(unprocessedEvents, processedEvents, errorChan, errorReports)

	ackable := emptyAckable{}
	deployEvent := listener.RegistryEvent{
		Action:  "INSERT",
		Digest:  app.RepoHost + "/" + app.RepoPath + "@" + sha,
		Tag:     app.ImageTag,
		Message: ackable,
	}

	unprocessedEvents <- &deployEvent

	select {
	case <-errorChan:
		close(unprocessedEvents)
		return fmt.Errorf("deploy failed")
	case <-processedEvents:
		close(unprocessedEvents)
		return nil
	}
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

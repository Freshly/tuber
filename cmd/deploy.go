package cmd

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/slack"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "deploy [app]",
	Short:        "deploys the latest built image of an app. CURRENTLY REQUIRES A LOCAL DB",
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

	creds, err := credentials()
	if err != nil {
		return err
	}

	// TODO: keep this available but under a flag, and move all the default behavior into a graphql mutation
	db, err := db()
	if err != nil {
		return err
	}
	defer db.Close()
	// err = pullLocalDB(db)
	// if err != nil {
	// 	return err
	// }

	app, err := db.App(appName)
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
	processor := events.NewProcessor(ctx, logger, db, creds, data, viper.GetBool("reviewapps-enabled"), slackClient)

	ref, err := name.ParseReference(app.ImageTag)
	if err != nil {
		return err
	}

	img, err := remote.Image(ref, remote.WithAuth(google.NewJSONKeyAuthenticator(string(creds))))
	if err != nil {
		return err
	}

	digest, err := img.Digest()
	if err != nil {
		return err
	}

	event, err := events.NewEvent(logger, ref.Context().Digest(digest.String()).String(), app.ImageTag)
	if err != nil {
		return fmt.Errorf("app image tag invalid")
	}

	processor.StartRelease(event, app)
	return nil
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

package cmd

import (
	"context"
	"tuber/pkg/core"
	"tuber/pkg/events"

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

	// location := app.GetRepositoryLocation()
	//
	// sha, err := containers.GetLatestSHA(location, creds)
	//
	// if err != nil {
	// 	return err
	// }

	data, err := clusterData()
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	processor := events.NewProcessor(ctx, logger, creds, data, viper.GetBool("reviewapps-enabled"))
	// digest := app.RepoHost + "/" + app.RepoPath + "@" + sha
	var digest string

	// broken
	// digest = "gcr.io/freshly-docker/potatoes@sha256:30ad0ca001a3e5994568ecde20aee13f33559dfd57a9f6bf2c986fbde147bccc"
	// working
	digest = "gcr.io/freshly-docker/potatoes@sha256:ec02e159543df71b718b664939bfe0eb0bb3e4d1850a3a1b5d355732661ebac1"
	// digest = "gcr.io/freshly-docker/potatoes@sha256:333f797a7fb5919f4b6cfc30de33cf48f73bfd8ca709ca7e4661612b104d4026"

	tag := app.ImageTag

	processor.ProcessMessage(digest, tag)
	return nil
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

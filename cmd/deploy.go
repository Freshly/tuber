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

	// pre new stuff
	// digest = "gcr.io/freshly-docker/potatoes@sha256:4d2138f4212707bc0033b8299bf7d567844e30c255cee1f78b3b0cc11a4e4283"
	// digest = "gcr.io/freshly-docker/potatoes@sha256:1aec3c61bed62c4483fa30d742d87f1e5f364751fe2da4a1bbed5d0cc92e1b27"

	// with new stuff
	// digest = "gcr.io/freshly-docker/potatoes@sha256:eef3d58e356a409398fcf4ca52441c7f0ebc501c4b282a2597327ba4b8dc3ec2"
	// digest = "gcr.io/freshly-docker/potatoes@sha256:7370466b40bf27c1f89039f4883ec9b9f9d193e3afdc7f59615ea10cafde34db"

	// broken
	digest = "gcr.io/freshly-docker/potatoes@sha256:30ad0ca001a3e5994568ecde20aee13f33559dfd57a9f6bf2c986fbde147bccc"
	// wtf
	// digest = "gcr.io/freshly-docker/potatoes@sha256:ec02e159543df71b718b664939bfe0eb0bb3e4d1850a3a1b5d355732661ebac1"
	// digest = "gcr.io/freshly-docker/potatoes@sha256:333f797a7fb5919f4b6cfc30de33cf48f73bfd8ca709ca7e4661612b104d4026"

	tag := app.ImageTag

	processor.ProcessMessage(digest, tag)
	return nil
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

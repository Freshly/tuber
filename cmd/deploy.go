package cmd

import (
	"context"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/gcr"
	"github.com/freshly/tuber/pkg/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deployCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "deploy [app]",
	Short:         "deploys the latest built image of an app, or a certain tag if specified",
	RunE:          deploy,
	PreRunE:       promptCurrentContext,
	Args:          cobra.ExactArgs(1),
}

func deploy(cmd *cobra.Command, args []string) error {
	appName := args[0]
	if deployLocalFlag {
		return localDeploy(appName, deployTagFlag)
	}

	tag := deployTagFlag
	if tag == "" {
		app, err := getApp(appName)
		if err != nil {
			return err
		}
		tag = app.ImageTag
	}

	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	gql := `
		mutation($input: DeployInput!) {
			deploy(input: $input) {
				name
			}
		}
	`

	input := &model.DeployInput{
		Name: appName,
		Tag:  &tag,
	}

	var respData struct {
		deploy *model.TuberApp
	}

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func localDeploy(appName string, flagTag string) error {
	logger, err := createLogger()
	if err != nil {
		return err
	}

	defer logger.Sync()

	creds, err := credentials()
	if err != nil {
		return err
	}

	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	remoteApp, err := getApp(appName)
	if err != nil {
		return err
	}
	err = db.SaveApp(remoteApp)
	if err != nil {
		return err
	}

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

	tag := flagTag
	if tag == "" {
		tag = app.ImageTag
	}

	digest, err := gcr.DigestFromTag(tag, creds)
	if err != nil {
		return err
	}

	processor.StartRelease(events.NewEvent(logger, digest, tag), app)
	return nil
}

var deployLocalFlag bool
var deployTagFlag string

func init() {
	deployCmd.Flags().BoolVar(&deployLocalFlag, "local", false, "run the full deploy process locally, including all monitoring.")
	deployCmd.Flags().StringVarP(&deployTagFlag, "tag", "t", "", "deploy a specific tag")
	rootCmd.AddCommand(deployCmd)
}

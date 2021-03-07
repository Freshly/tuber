package cmd

import (
	"context"
	"fmt"
	"tuber/pkg/core"
	"tuber/pkg/k8s"
	"tuber/pkg/proto"
	"tuber/pkg/reviewapps"

	"github.com/spf13/cobra"
)

var reviewAppsCmd = &cobra.Command{
	Use:   "review-apps [command]",
	Short: "A root command for review app configurating",
}

var reviewAppsCreateCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "create [source app name] [branch name]",
	Short:        "",
	Args:         cobra.ExactArgs(2),
	RunE:         create,
}

var reviewAppsDeleteCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "delete [review app name]",
	Short:        "",
	Args:         cobra.ExactArgs(1),
	RunE:         delete,
}

func create(cmd *cobra.Command, args []string) error {
	sourceAppName := args[0]
	branch := args[1]

	tuberConf, err := getTuberConfig()
	if err != nil {
		return err
	}

	clusterConf := tuberConf.CurrentClusterConfig()
	if clusterConf.URL == "" {
		return fmt.Errorf("no tuber url found for current cluster. run `tuber config`")
	}

	client, conn, err := reviewapps.NewClient(clusterConf.URL)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = core.FindApp(sourceAppName)
	if err != nil {
		return fmt.Errorf("source app not found")
	}

	canDeploy, err := k8s.CanDeploy(sourceAppName)
	if err != nil {
		return err
	}
	if !canDeploy {
		return fmt.Errorf("not permitted to create a review app from %s", sourceAppName)
	}

	config, err := k8s.GetConfig()
	if err != nil {
		return err
	}

	req := proto.CreateReviewAppRequest{
		AppName: sourceAppName,
		Branch:  branch,
		Token:   config.AccessToken,
	}

	res, err := client.CreateReviewApp(context.Background(), &req)
	if err != nil {
		return err
	}

	if res.Error != "" {
		return fmt.Errorf(res.Error)
	}

	fmt.Println("Created review app")
	fmt.Println(res.Hostname)

	return nil
}

func delete(cmd *cobra.Command, args []string) error {
	reviewAppName := args[0]

	tuberConf, err := getTuberConfig()
	if err != nil {
		return err
	}

	clusterConf := tuberConf.CurrentClusterConfig()
	if clusterConf.URL == "" {
		return fmt.Errorf("no tuber url found for current cluster. run `tuber config`")
	}

	client, conn, err := reviewapps.NewClient(clusterConf.URL)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = core.FindReviewApp(reviewAppName)
	if err != nil {
		return fmt.Errorf("review app not found")
	}

	config, err := k8s.GetConfig()
	if err != nil {
		return err
	}

	req := proto.DeleteReviewAppRequest{
		AppName: reviewAppName,
		Token:   config.AccessToken,
	}

	res, err := client.DeleteReviewApp(context.Background(), &req)
	if err != nil {
		return err
	}

	if res.Error != "" {
		return fmt.Errorf(res.Error)
	}

	fmt.Println("Deleted review app")

	return nil
}

func init() {
	rootCmd.AddCommand(reviewAppsCmd)
	reviewAppsCmd.AddCommand(reviewAppsCreateCmd)
	reviewAppsCmd.AddCommand(reviewAppsDeleteCmd)
}

package cmd

import (
	"context"
	"fmt"
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

func create(cmd *cobra.Command, args []string) error {
	sourceAppName := args[0]
	branch := args[1]

	tuberConf, err := getTuberConfig()
	if err != nil {
		return err
	}

	clusterConf := tuberConf.CurrentClusterConfig()
	if clusterConf.URL == "" {
		return fmt.Errorf("no cluster config found. run `tuber config`")
	}

	client, conn := reviewapps.NewClient(clusterConf.URL)
	defer conn.Close()

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

func init() {
	rootCmd.AddCommand(reviewAppsCmd)
	reviewAppsCmd.AddCommand(reviewAppsCreateCmd)
}

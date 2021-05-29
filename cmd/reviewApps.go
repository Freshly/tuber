package cmd

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/reviewapps"

	"github.com/spf13/cobra"
)

var reviewAppsCmd = &cobra.Command{
	Use:   "review-apps [command]",
	Short: "A root command for review app configurating",
}

var reviewAppsCreateCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "create [source app name] [branch name]",
	Short:        "Create a temporary application deployed alongside the source application for a given branch, copying its rolebindings and env",
	Args:         cobra.ExactArgs(2),
	RunE:         create,
}

var reviewAppsDeleteCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "delete [source app name] [branch]",
	Short:        "Delete a review app",
	Args:         cobra.ExactArgs(2),
	RunE:         delete,
}

func create(cmd *cobra.Command, args []string) error {
	sourceAppName := args[0]
	branchName := args[1]

	canDeploy, err := k8s.CanDeploy(sourceAppName)
	if err != nil {
		return err
	}
	if !canDeploy {
		return fmt.Errorf("not permitted to create a review app from %s", sourceAppName)
	}

	graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)

	appName := args[0]

	input := &model.CreateReviewAppInput{
		Name:       appName,
		BranchName: branchName,
	}

	var respData struct {
		destoryApp *model.TuberApp
	}

	gql := `
		mutation($input: CreateReviewAppInput!) {
			createReviewApp(input: $input) {
				name
			}
		}
	`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func delete(cmd *cobra.Command, args []string) error {
	sourceAppName := args[0]
	branch := args[1]
	reviewAppName := reviewapps.ReviewAppName(sourceAppName, branch)
	return destroyApp(cmd, []string{reviewAppName})
}

func init() {
	rootCmd.AddCommand(reviewAppsCmd)
	reviewAppsCmd.AddCommand(reviewAppsCreateCmd)
	reviewAppsCmd.AddCommand(reviewAppsDeleteCmd)
}

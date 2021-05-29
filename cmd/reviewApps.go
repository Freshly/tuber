package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/reviewapps"
	"github.com/olekukonko/tablewriter"

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

var reviewAppsListCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "list [source app name]",
	Short:        "Delete a review app",
	Args:         cobra.ExactArgs(1),
	RunE:         listReviewApps,
}

func listReviewApps(cmd *cobra.Command, args []string) (err error) {
	graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)
	appName := args[0]

	gql := `
			query($name: String!) {
				getApp(name: $name) {
					name

					reviewApps {
						name
						imageTag
					}
				}
			}
		`

	var respData struct {
		GetApp *model.TuberApp
	}

	if err := graphql.Query(context.Background(), gql, &respData, graph.WithVar("name", appName)); err != nil {
		return err
	}

	apps := respData.GetApp.ReviewApps

	sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })

	if jsonOutput {
		out, err := json.Marshal(apps)

		if err != nil {
			return err
		}

		os.Stdout.Write(out)

		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Image"})
	table.SetBorder(false)

	for _, app := range apps {
		table.Append([]string{app.Name, app.ImageTag})
	}

	table.Render()
	return nil
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
	reviewAppsCmd.AddCommand(reviewAppsListCmd)
}

package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var reviewAppReaperCmd = &cobra.Command{
	SilenceUsage: true,
	Hidden:       true,
	Use:          "review-app-reaper",
	Short:        "very internal, deletes old review apps",
	Args:         cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !viper.GetBool("TUBER_REVIEWAPPS_ENABLED") {
			return nil
		}
		out, err := k8s.GetCollection("secrets", "tuber", `-o=jsonpath='{.items[?(@.metadata.annotations.kubernetes\.io/service-account\.name=="tuber")].data.token}'`)
		if err != nil {
			return err
		}

		decoded, err := base64.StdEncoding.DecodeString(strings.Trim(string(out), "\r\n'"))
		if err != nil {
			return err
		}

		client := graph.NewClient("http://tuber.tuber:3000", viper.GetString("TUBER_OAUTH_WEB_CLIENT_ID"))
		client.IntraCluster = true
		client.IntraClusterToken = string(decoded)

		var allReviewApps struct {
			GetAllReviewApps []*model.TuberApp
		}
		gql := `
			query {
				getAllReviewApps {
					name
					createdAt
					updatedAt
				}
			}
		`
		err = client.Query(context.Background(), gql, &allReviewApps)
		if err != nil {
			return err
		}
		for _, app := range allReviewApps.GetAllReviewApps {
			updatedAt, err := app.ParsedUpdatedAt()
			if err != nil {
				return fmt.Errorf("error parsing updated at for app: %s, %v", app.Name, err)
			}
			if time.Since(updatedAt).Hours() > 48 {
				fmt.Println("destroying " + app.Name)
				input := &model.AppInput{
					Name: app.Name,
				}

				var destroyResponse struct {
					destroyApp *model.TuberApp
				}

				gql := `
					mutation($input: AppInput!) {
						destroyApp(input: $input) {
							name
						}
					}
				`
				err = client.Mutation(context.Background(), gql, nil, input, &destroyResponse)
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewAppReaperCmd)
}

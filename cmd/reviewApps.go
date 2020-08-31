/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"tuber/pkg/client"
	"tuber/pkg/k8s"
	"tuber/pkg/proto"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
)

// reviewAppsCmd represents the reviewApps command
var reviewAppsCmd = &cobra.Command{
	Use:   "review-apps",
	Short: "A brief description of your command",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("reviewApps called")
	},
}

var reviewAppsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long:  "",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]
		branch := args[1]

		c, conn := client.NewClient()
		defer conn.Close()

		auth, err := k8s.ClusterToken()
		if err != nil {
			return err
		}

		req := proto.Request{
			AppName: appName,
			Branch:  branch,
			Token:   auth,
		}

		res, err := c.CreateReviewApp(context.Background(), &req)

		if err != nil {
			return err
		}

		spew.Dump(res)

		fmt.Println("reviewApps called")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewAppsCmd)
	reviewAppsCmd.AddCommand(reviewAppsCreateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reviewAppsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reviewAppsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

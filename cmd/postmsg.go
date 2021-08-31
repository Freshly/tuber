package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

type Message struct {
	AppName   string
	CommitSha string
	Repo      string
	Branch    string
}

var postMsg = &cobra.Command{
	SilenceUsage: true,
	Use:          "postMessage",
	Short:        "post message to pubsub",
	Args:         cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		creds, err := credentials()
		if err != nil {
			panic(err)
		}

		client, err := pubsub.NewClient(ctx, "freshly-docker", option.WithCredentialsJSON(creds))
		if err != nil {
			return err
		}

		fmt.Println("client.Topic")
		topic := client.Topic("tuber-test-topic")

		msg := Message{
			AppName:   "testyapp",
			CommitSha: "woogabooga",
			Repo:      "repo",
			Branch:    "branch",
		}
		marshalled, err := json.Marshal(&msg)
		if err != nil {
			return err
		}

		fmt.Println("topic.Publish")
		res := topic.Publish(ctx, &pubsub.Message{Data: marshalled})
		_, err = res.Get(ctx)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(postMsg)
}

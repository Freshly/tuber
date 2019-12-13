package listen

import (
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// RegistryEvent json deserialize target for pubsub
type RegistryEvent struct {
	Action string `json:"action"`
	Digest string `json:"digest"`
}

type callback func(*RegistryEvent, error)

// Listen it listens
func Listen(listener callback) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "freshly-docker", option.WithCredentialsFile("./credentials.json"))

	if err != nil {
		log.Fatal(err) // Error will always be not nil. and not always an error.
	}

	subscription := client.Subscription("freshly-docker-gcr-events")

	err = subscription.Receive(context.Background(), func(ctx context.Context, message *pubsub.Message) {
		var obj = new(RegistryEvent)
		err := json.Unmarshal(message.Data, obj)

		listener(obj, err)
	})

	return err
}

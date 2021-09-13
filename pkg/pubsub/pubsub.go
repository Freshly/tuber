package pubsub

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/report"

	"cloud.google.com/go/pubsub"
	"go.uber.org/zap"

	"google.golang.org/api/option"
)

// Listener is a pubsub server that pipes messages off to its events.Processor
type Listener struct {
	ctx              context.Context
	logger           *zap.Logger
	pubsubProject    string
	subscriptionName string
	credentials      []byte
	clusterData      *core.ClusterData
	processor        Processor
}

type Processor interface {
	ProcessMessage(Message)
}

// NewListener is a constructor for Listener with field validation
func NewListener(ctx context.Context, logger *zap.Logger, pubsubProject string, subscriptionName string,
	credentials []byte, clusterData *core.ClusterData, processor Processor) (*Listener, error) {
	if logger == nil {
		return nil, errors.New("zap logger is required")
	}
	if pubsubProject == "" {
		return nil, errors.New("pubsub project is required")
	}
	if subscriptionName == "" {
		return nil, errors.New("pubsub subscription name is required")
	}

	return &Listener{
		ctx:              ctx,
		logger:           logger,
		pubsubProject:    pubsubProject,
		subscriptionName: subscriptionName,
		credentials:      credentials,
		clusterData:      clusterData,
		processor:        processor,
	}, nil
}

// Message json deserialization target for pubsub messages
type Message struct {
	// These attributes are for GCR Messages
	Digest string `json:"digest"`
	Tag    string `json:"tag"`

	// These are for Cloud Build Notifications.
	Status string   `json:"status"`
	Images []string `json:"images"`
	LogURL string   `json:"logUrl"`
}

// Start starts up the pubsub server and pipes incoming messages to the Listener's events.Processor
func (l *Listener) Start() error {
	client, err := pubsub.NewClient(l.ctx, l.pubsubProject, option.WithCredentialsJSON(l.credentials))
	if err != nil {
		client, err = pubsub.NewClient(l.ctx, l.pubsubProject)
	}

	listenLogger := l.logger.With(zap.String("context", "pubsubServer"))
	if err != nil {
		listenLogger.With(zap.Error(err)).Warn("pubsub client initialization failed")
		report.Error(err, report.Scope{"context": "pubsub client initialization"})
		return err
	}

	subscription := client.Subscription(l.subscriptionName)

	listenLogger.Debug("pubsub server starting")
	listenLogger.Debug("subscription options", zap.Reflect("options", subscription.ReceiveSettings))

	err = subscription.Receive(l.ctx, func(ctx context.Context, pubsubMessage *pubsub.Message) {
		pubsubMessage.Ack()
		// {"action":"INSERT","digest":"gcr.io/freshly-docker/freshly@sha256:17f4431497a07da98bc16e599ef9d38afb9817049b6e98b71b7e321b946a24d4",
		// "tag":"gcr.io/freshly-docker/freshly:PIG-267-refactor-email-service"}

		var message Message
		err = json.Unmarshal(pubsubMessage.Data, &message)
		if err != nil {
			listenLogger.Warn("failed to unmarshal pubsub message", zap.Error(err))
			report.Error(err, report.Scope{"context": "messageProcessing"})
			return
		}

		l.processor.ProcessMessage(message)
	})

	if err != nil {
		listenLogger.With(zap.Error(err)).Warn("pubsub listener halted")
		report.Error(err, report.Scope{"context": "pubsub listener halted"})
	}
	listenLogger.Debug("listener stopped")

	return err
}

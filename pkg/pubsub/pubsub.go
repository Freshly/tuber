package pubsub

import (
	"context"
	"errors"
	"tuber/pkg/core"
	"tuber/pkg/events"
	"tuber/pkg/report"

	"go.uber.org/zap"

	cloudpubsub "cloud.google.com/go/pubsub"
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
	processor        events.Processor
}

// Message is an alias for cloud pubsub message
type Message = *cloudpubsub.Message

// NewListener is a constructor for Listener with field validation
func NewListener(ctx context.Context, logger *zap.Logger, pubsubProject string, subscriptionName string,
	credentials []byte, clusterData *core.ClusterData, processor events.Processor) (*Listener, error) {
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

// Listen starts up the pubsub server and pipes incoming pubsub to the Listener's events.Processor
func (l *Listener) Listen() error {
	var client *cloudpubsub.Client
	var err error

	client, err = cloudpubsub.NewClient(l.ctx, l.pubsubProject, option.WithCredentialsJSON(l.credentials))

	if err != nil {
		client, err = cloudpubsub.NewClient(l.ctx, l.pubsubProject)
	}

	if err != nil {
		return err
	}

	subscription := client.Subscription(l.subscriptionName)

	listenLogger := l.logger.With(zap.String("context", "pubsubServer"))
	listenLogger.Debug("pubsub server starting")
	listenLogger.Debug("subscription options", zap.Reflect("options", subscription.ReceiveSettings))

	err = subscription.Receive(l.ctx, func(ctx context.Context, message *cloudpubsub.Message) {
		message.Ack()
		go l.processor.ProcessMessage(&message)
	})

	if err != nil {
		listenLogger.With(zap.Error(err)).Warn("receiver error")
		report.Error(err, report.Scope{"context": "pubsubServer"})
	}
	listenLogger.Debug("listener stopped")

	return err
}

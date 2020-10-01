package listener

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

//listener binds to a pubsub subscription and sends messages to a queue
type listener struct {
	projectID    string
	subscription string

	unprocessed    chan *RegistryEvent
	processed      chan *RegistryEvent
	failedReleases chan FailedRelease

	// TODO: What should this channel receive?
	reportableErrors chan error

	wait *sync.WaitGroup

	logger       *zap.Logger
	recvSettings pubsub.ReceiveSettings
}

// Option provides optional settings to a listener constructor
type Option func(*listener)

// WithMaxAccept determines the maximum number of outstanding messages accepted
func WithMaxAccept(n int) Option {
	return func(l *listener) {
		l.recvSettings.MaxOutstandingMessages = n
	}
}

// WithMaxTimeout sets the maximum ack timeout extension
func WithMaxTimeout(d time.Duration) Option {
	return func(l *listener) {
		l.recvSettings.MaxExtension = d
	}
}

// NewListener creates a new PubSub listener
func NewListener(logger *zap.Logger, subscriptionName string, options ...Option) *listener {
	var l = &listener{
		projectID:    "freshly-docker",
		subscription: subscriptionName,

		unprocessed:      make(chan *RegistryEvent, 1),
		processed:        make(chan *RegistryEvent, 1),
		failedReleases:   make(chan FailedRelease, 1),
		reportableErrors: make(chan error, 1),
		wait:             &sync.WaitGroup{},
		logger:           logger,
		recvSettings:     pubsub.ReceiveSettings{},
	}

	for _, op := range options {
		op(l)
	}
	return l
}

// Listen for incoming pubsub requests
func (l *listener) Listen(ctx context.Context, credentials []byte) (<-chan *RegistryEvent, chan<- *RegistryEvent, chan<- FailedRelease, error) {
	go l.startAcker(ctx)

	var err = l.startListener(ctx, credentials)
	return l.unprocessed, l.processed, l.failedReleases, err
}

func (l *listener) startListener(ctx context.Context, credentials []byte) error {
	var client *pubsub.Client
	var err error

	client, err = pubsub.NewClient(ctx, l.projectID, option.WithCredentialsJSON(credentials))

	if err != nil {
		client, err = pubsub.NewClient(ctx, l.projectID)
	}

	if err != nil {
		return err
	}

	subscription := client.Subscription(l.subscription)
	subscription.ReceiveSettings = l.recvSettings

	go func(in chan<- *RegistryEvent, logger *zap.Logger) {
		// Register this goroutine in the waiter
		l.wait.Add(1)
		defer l.wait.Done()

		// Close the message channel before exiting to signal to downstream that we're done
		defer close(in)

		l.logger.Debug("listener: starting")
		l.logger.Debug("listener: subscription options", zap.Reflect("options", subscription.ReceiveSettings))
		err = subscription.Receive(ctx,
			func(ctx context.Context, message *pubsub.Message) {
				obj := &RegistryEvent{Message: message}
				jsonErr := json.Unmarshal(message.Data, obj)

				if jsonErr != nil {
					l.logger.Warn("could not unmarshal message")
				} else {
					l.logger.Debug("Sending event to unprocessed channel from listener", zap.String("tag", obj.Tag), zap.String("digest", obj.Digest))
					in <- obj
				}
			})

		if err != nil {
			l.logger.With(zap.Error(err)).Warn("listener: receiver error")
		}
		l.logger.Debug("listener: shutting down")
	}(l.unprocessed, l.logger)

	return err
}

func (l *listener) startAcker(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	// Register this goroutine in the waiter
	l.wait.Add(1)
	defer l.wait.Done()

	ackLogger := l.logger.With(
		zap.String("action", "acknowledger"),
	)

	ackLogger.Debug("starting")

	for event := range l.processed {
		ackLogger.Debug("Acknowledging",
			zap.String("tag", event.Tag),
			zap.String("digest", event.Digest),
		)
		event.Message.Ack()
		ackLogger.Info("Acknowledged",
			zap.String("tag", event.Tag),
			zap.String("digest", event.Digest),
		)
	}

	ackLogger.Debug("stopped")
}

func (l *listener) startNacker(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	l.wait.Add(1)
	defer l.wait.Done()

	l.logger.Debug("error loop: starting")

	for failure := range l.failedReleases {
		l.logger.Debug("nacking",
			zap.String("tag", failure.Event.Tag),
			zap.String("digest", failure.Event.Digest),
		)

		l.logger.Warn("failed release", zap.Error(failure.Err))
		failure.Event.Message.Nack()

		l.logger.Info("nacked",
			zap.String("tag", failure.Event.Tag),
			zap.String("digest", failure.Event.Digest),
		)
	}

	l.logger.Debug("error loop: stopped")
}

// Wait for the listener and acker goroutines to exit.
// If you use this method, you must ensure that you close
// the output channel when no more work is being processed
func (l *listener) Wait() {
	l.wait.Wait()
}

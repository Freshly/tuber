package events

import (
	"context"
	"encoding/json"
	"time"
	"tuber/pkg/containers"
	"tuber/pkg/core"
	"tuber/pkg/report"

	"cloud.google.com/go/pubsub"
	"go.uber.org/zap"
)

// Processor processes events
type Processor struct {
	ctx               context.Context
	logger            *zap.Logger
	creds             []byte
	clusterData       *core.ClusterData
	reviewAppsEnabled bool
}

func NewProcessor(ctx context.Context, logger *zap.Logger, creds []byte, clusterData *core.ClusterData, reviewAppsEnabled bool) Processor {
	return Processor{
		ctx:               ctx,
		logger:            logger,
		creds:             creds,
		clusterData:       clusterData,
		reviewAppsEnabled: reviewAppsEnabled,
	}
}

// Event json deserialize target for pubsub
type Event struct {
	Action     string `json:"action"`
	Digest     string `json:"digest"`
	Tag        string `json:"tag"`
	message    *pubsub.Message
	logger     *zap.Logger
	errorScope report.Scope
}

func (p Processor) ProcessMessage(message *pubsub.Message) {
	event, err := p.eventFromMessage(message)
	if err != nil {
		p.logger.Warn("could not unmarshal pubsub message", zap.Error(err))
		report.Error(err, report.Scope{"context": "messageProcessing"})
		return
	}

	apps, err := p.apps()
	if err != nil {
		event.logger.Error("could not look up tuber apps", zap.Error(err))
		report.Error(err, event.errorScope.WithContext("tuber apps lookup"))
		return
	}

	p.processEvent(event, apps)
}

func (p Processor) eventFromMessage(message *pubsub.Message) (*Event, error) {
	event := &Event{message: message}
	err := json.Unmarshal(message.Data, event)
	if err != nil {
		return nil, err
	}
	event.logger = p.logger.With(zap.String("tag", event.Tag), zap.String("digest", event.Digest))
	event.errorScope = report.Scope{"tag": event.Tag, "digest": event.Digest}
	return event, nil
}

func (p Processor) apps() ([]core.TuberApp, error) {
	if p.reviewAppsEnabled {
		p.logger.Debug("listing source and review apps")
		return core.SourceAndReviewApps()
	}

	p.logger.Debug("listing source apps")
	return core.TuberSourceApps()
}

func (p Processor) processEvent(event *Event, apps []core.TuberApp) {
	event.logger.Debug("processing event")
	event.logger.Debug("current tuber apps", zap.Any("apps", apps))

	matchFound := false
	for _, app := range apps {
		if app.ImageTag == event.Tag {
			matchFound = true
			p.deploy(event, &app)
		}
	}
	if !matchFound {
		event.logger.Info("ignored event")
	}
}

func (p Processor) deploy(event *Event, app *core.TuberApp) {
	deployLogger := event.logger.With(
		zap.String("name", app.Name),
		zap.String("branch", app.Tag),
		zap.String("imageTag", app.ImageTag),
		zap.String("action", "release"),
	)
	errorScope := event.errorScope.AddScope(report.Scope{
		"name":     app.Name,
		"branch":   app.Tag,
		"imageTag": app.ImageTag,
	})

	deployLogger.Info("release starting")

	startTime := time.Now()
	prereleaseYamls, releaseYamls, err := containers.GetTuberLayer(app.GetRepositoryLocation(), p.creds)

	if err != nil {
		deployLogger.Warn("failed to get tuber layer", zap.Error(err))
		report.Error(err, errorScope.WithContext("get tuber layer"))
		return
	}

	if len(prereleaseYamls) > 0 {
		deployLogger.Info("prerelease starting")

		err = core.RunPrerelease(prereleaseYamls, app, event.Digest, p.clusterData)

		if err != nil {
			report.Error(err, errorScope.WithContext("prerelease"))
			deployLogger.Warn("failed prerelease", zap.Error(err))
			return
		}

		deployLogger.Info("prerelease complete")
	}

	releaseIDs, err := core.ReleaseTubers(releaseYamls, app, event.Digest, p.clusterData)
	if err != nil {
		deployLogger.Warn("failed release", zap.Error(err))
		report.Error(err, errorScope.WithContext("release"))
		return
	}
	deployLogger.Info("release complete", zap.Strings("releaseIds", releaseIDs), zap.Duration("duration", time.Since(startTime)))

	return
}

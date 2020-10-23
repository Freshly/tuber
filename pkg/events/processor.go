package events

import (
	"context"
	"time"
	"tuber/pkg/containers"
	"tuber/pkg/core"
	"tuber/pkg/report"

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

// NewProcessor is a constructor for Processors so that the fields can be unexported
func NewProcessor(ctx context.Context, logger *zap.Logger, creds []byte, clusterData *core.ClusterData, reviewAppsEnabled bool) Processor {
	return Processor{
		ctx:               ctx,
		logger:            logger,
		creds:             creds,
		clusterData:       clusterData,
		reviewAppsEnabled: reviewAppsEnabled,
	}
}

type event struct {
	digest     string
	tag        string
	logger     *zap.Logger
	errorScope report.Scope
}

// ProcessMessage receives a pubsub message, filters it against TuberApps, and triggers deploys for matching apps
func (p Processor) ProcessMessage(digest string, tag string) {
	event := event{
		digest:     digest,
		tag:        tag,
		logger:     p.logger.With(zap.String("tag", tag), zap.String("digest", digest)),
		errorScope: report.Scope{"tag": tag, "digest": digest},
	}
	apps, err := p.apps()
	if err != nil {
		event.logger.Error("failed to look up tuber apps", zap.Error(err))
		report.Error(err, event.errorScope.WithContext("tuber apps lookup"))
		return
	}
	event.logger.Debug("filtering event against current tuber apps", zap.Any("apps", apps))

	matchFound := false
	for _, app := range apps {
		if app.ImageTag == event.tag {
			matchFound = true
			p.deploy(event, &app)
		}
	}
	if !matchFound {
		event.logger.Debug("ignored event")
	}
}

func (p Processor) apps() ([]core.TuberApp, error) {
	if p.reviewAppsEnabled {
		p.logger.Debug("pulling source and review apps")
		return core.SourceAndReviewApps()
	}

	p.logger.Debug("pulling source apps")
	return core.TuberSourceApps()
}

func (p Processor) deploy(event event, app *core.TuberApp) {
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
		deployLogger.Warn("failed to find tuber layer", zap.Error(err))
		report.Error(err, errorScope.WithContext("find tuber layer"))
		return
	}

	if len(prereleaseYamls) > 0 {
		deployLogger.Info("prerelease starting")

		err = core.RunPrerelease(prereleaseYamls, app, event.digest, p.clusterData)

		if err != nil {
			report.Error(err, errorScope.WithContext("prerelease"))
			deployLogger.Warn("failed prerelease", zap.Error(err))
			return
		}

		deployLogger.Info("prerelease complete")
	}

	releaseIDs, err := core.ReleaseTubers(releaseYamls, app, event.digest, p.clusterData)
	if err != nil {
		deployLogger.Warn("failed release", zap.Error(err))
		report.Error(err, errorScope.WithContext("release"))
		return
	}
	deployLogger.Info("release complete", zap.Strings("releaseIds", releaseIDs), zap.Duration("duration", time.Since(startTime)))

	return
}

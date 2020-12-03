package events

import (
	"context"
	"fmt"
	"strings"
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
	sha        string
}

// ProcessMessage receives a pubsub message, filters it against TuberApps, and triggers releases for matching apps
func (p Processor) ProcessMessage(digest string, tag string) {
	logger := p.logger.With(zap.String("tag", tag), zap.String("digest", digest))
	scope := report.Scope{"tag": tag, "digest": digest}
	split := strings.Split(digest, "@")
	if len(split) != 2 {
		err := fmt.Errorf("event digest split length not 2")
		logger.Error("failed to process event", zap.Error(err))
		report.Error(err, scope.WithContext("event processing"))
		return
	}
	sha := split[1]

	event := event{
		digest:     digest,
		tag:        tag,
		logger:     logger,
		errorScope: scope,
		sha:        sha,
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

			paused, err := core.ReleasesPaused(app.Name)
			if err != nil {
				event.logger.Error("failed to check for paused state", zap.Error(err))
			}

			if paused {
				event.logger.Warn("app deployments paused; skipping", zap.String("appName", app.Name))
				continue
			}

			p.startRelease(event, &app)
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

func (p Processor) startRelease(event event, app *core.TuberApp) {
	logger := event.logger.With(
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

	logger.Info("release starting")
	prereleaseYamls, releaseYamls, err := containers.GetTuberLayer(app.GetRepositoryLocation(), event.sha, p.creds)
	if err != nil {
		logger.Error("failed to find tuber layer", zap.Error(err))
		report.Error(err, errorScope.WithContext("find tuber layer"))
		return
	}

	if len(prereleaseYamls) > 0 {
		logger.Info("prerelease starting")

		err = core.RunPrerelease(prereleaseYamls, app, event.digest, p.clusterData)
		if err != nil {
			report.Error(err, errorScope.WithContext("prerelease"))
			logger.Error("failed prerelease", zap.Error(err))
			return
		}

		logger.Info("prerelease complete")
	}

	startTime := time.Now()
	err = core.Release(logger, errorScope, releaseYamls, app, event.digest, p.clusterData)
	if err != nil {
		logger.Warn("release failed", zap.Error(err), zap.Duration("duration", time.Since(startTime)))
		return
	}
	logger.Info("release complete", zap.Duration("duration", time.Since(startTime)))
	return
}

package events

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/containers"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/report"
	"github.com/freshly/tuber/pkg/slack"

	"go.uber.org/zap"
)

// Processor processes events
type Processor struct {
	ctx               context.Context
	logger            *zap.Logger
	creds             []byte
	clusterData       *core.ClusterData
	reviewAppsEnabled bool
	locks             *map[string]*sync.Cond
	slackClient       *slack.Client
}

// NewProcessor is a constructor for Processors so that the fields can be unexported
func NewProcessor(ctx context.Context, logger *zap.Logger, creds []byte, clusterData *core.ClusterData, reviewAppsEnabled bool, slackClient *slack.Client) Processor {
	l := make(map[string]*sync.Cond)

	return Processor{
		ctx:               ctx,
		logger:            logger,
		creds:             creds,
		clusterData:       clusterData,
		reviewAppsEnabled: reviewAppsEnabled,
		locks:             &l,
		slackClient:       slackClient,
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
		logger.Warn("failed to process event", zap.Error(fmt.Errorf("event digest split length not 2")))
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

	wg := sync.WaitGroup{}

	for _, a := range apps {
		if a.ImageTag == event.tag {
			matchFound = true
			wg.Add(1)

			go func(app model.TuberApp) {
				defer wg.Done()
				if _, ok := (*p.locks)[app.Name]; !ok {
					var mutex sync.Mutex
					(*p.locks)[app.Name] = sync.NewCond(&mutex)
				}

				cond := (*p.locks)[app.Name]
				cond.L.Lock()

				paused, err := core.ReleasesPaused(app.Name)
				if err != nil {
					event.logger.Error("failed to check for paused state", zap.Error(err))
					return
				}

				if paused {
					p.slackClient.Message(event.logger, "release skipped for "+app.Name+" as it is paused")
					event.logger.Warn("deployments are paused for this app; skipping", zap.String("appName", app.Name))
					return
				}
				p.startRelease(event, &app)
				cond.L.Unlock()
				cond.Signal()
			}(a)
		}
	}

	if !matchFound {
		event.logger.Debug("ignored event")
	} else {
		wg.Wait()
	}
}

func (p Processor) apps() ([]model.TuberApp, error) {
	if p.reviewAppsEnabled {
		p.logger.Debug("pulling source and review apps")
		return core.SourceAndReviewApps()
	}

	p.logger.Debug("pulling source apps")
	return core.TuberSourceApps()
}

func (p Processor) startRelease(event event, app *model.TuberApp) {
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

	yamls, err := containers.GetTuberLayer(core.GetRepositoryLocation(app), event.sha, p.creds)
	if err != nil {
		p.slackClient.Message(logger, "image not found for "+app.Name)
		logger.Error("failed to find tuber layer", zap.Error(err))
		report.Error(err, errorScope.WithContext("find tuber layer"))
		return
	}

	startTime := time.Now()
	err = core.Release(
		yamls,
		logger,
		errorScope,
		app,
		event.digest,
		p.clusterData,
	)

	if err != nil {
		logger.Warn("release failed", zap.Error(err), zap.Duration("duration", time.Since(startTime)))
		p.slackClient.Message(logger, ":loudspeaker: release failed for "+app.Name)
		return
	}

	p.slackClient.Message(logger, ":white_check_mark: release complete for "+app.Name)
	logger.Info("release complete", zap.Duration("duration", time.Since(startTime)))
	return
}

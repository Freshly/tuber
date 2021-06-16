package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/gcr"
	"github.com/freshly/tuber/pkg/report"
	"github.com/freshly/tuber/pkg/slack"
	"github.com/getsentry/sentry-go"

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
	db                *core.DB
}

// NewProcessor constructs a Processor
func NewProcessor(ctx context.Context, logger *zap.Logger, db *core.DB, creds []byte, clusterData *core.ClusterData, reviewAppsEnabled bool, slackClient *slack.Client) *Processor {
	l := make(map[string]*sync.Cond)

	return &Processor{
		ctx:               ctx,
		logger:            logger,
		creds:             creds,
		clusterData:       clusterData,
		reviewAppsEnabled: reviewAppsEnabled,
		locks:             &l,
		slackClient:       slackClient,
		db:                db,
	}
}

type Event struct {
	digest     string
	tag        string
	logger     *zap.Logger
	errorScope report.Scope
}

func NewEvent(logger *zap.Logger, digest string, tag string) *Event {
	logger = logger.With(zap.String("tag", tag), zap.String("digest", digest))
	scope := report.Scope{"tag": tag, "digest": digest}
	return &Event{
		digest:     digest,
		tag:        tag,
		logger:     logger,
		errorScope: scope,
	}
}

// ProcessMessage receives a pubsub message, filters it against TuberApps, and triggers releases for matching apps
func (p Processor) ProcessMessage(event *Event) {
	apps, err := p.db.AppsForTag(event.tag)
	if err != nil {
		event.logger.Error("failed to look up tuber apps", zap.Error(err))
		report.Error(err, event.errorScope.WithContext("tuber apps lookup"))
		return
	}

	if len(apps) == 0 {
		event.logger.Debug("ignored event")
		return
	}

	wg := sync.WaitGroup{}

	for _, a := range apps {
		wg.Add(1)

		go func(app *model.TuberApp) {
			defer sentry.Recover()
			defer wg.Done()
			p.ReleaseApp(event, app)
		}(a)
	}
	wg.Wait()
}

func (p Processor) ReleaseApp(event *Event, app *model.TuberApp) {
	// todo: the one in start _does not help mid-release panics_, errors package needs this functionality
	defer sentry.Recover()

	if _, ok := (*p.locks)[app.Name]; !ok {
		var mutex sync.Mutex
		(*p.locks)[app.Name] = sync.NewCond(&mutex)
	}

	cond := (*p.locks)[app.Name]
	cond.L.Lock()

	if app.Paused {
		p.slackClient.Message(event.logger, "release skipped for "+app.Name+" as it is paused")
		event.logger.Warn("deployments are paused for this app; skipping", zap.String("appName", app.Name))
		return
	}
	p.StartRelease(event, app)
	cond.L.Unlock()
	cond.Signal()
}

func (p Processor) StartRelease(event *Event, app *model.TuberApp) {
	logger := event.logger.With(
		zap.String("name", app.Name),
		zap.String("imageTag", app.ImageTag),
		zap.String("action", "release"),
	)

	errorScope := event.errorScope.AddScope(report.Scope{
		"name":     app.Name,
		"imageTag": app.ImageTag,
	})

	logger.Info("release starting")

	yamls, err := gcr.GetTuberLayer(event.digest, p.creds)
	if err != nil {
		p.slackClient.Message(logger, "image or tuber layer not found for "+app.Name)
		logger.Error("failed to find tuber layer", zap.Error(err))
		report.Error(err, errorScope.WithContext("find tuber layer"))
		return
	}

	var diffText string
	if app.GithubURL != "" && len(app.CurrentTags) != 0 {
		diffText = makeDiffText(app, yamls)
	}

	startTime := time.Now()
	err = core.Release(
		p.db,
		yamls,
		logger,
		errorScope,
		app,
		event.digest,
		p.clusterData,
		p.slackClient,
		diffText,
	)

	if err != nil {
		logger.Warn("release failed", zap.Error(err), zap.Duration("duration", time.Since(startTime)))
		p.slackClient.Message(logger, "<!here> :loudspeaker: release failed for "+app.Name)
		return
	}

	p.slackClient.Message(logger, ":white_check_mark: release complete for "+app.Name)
	logger.Info("release complete", zap.Duration("duration", time.Since(startTime)))
}

func makeDiffText(app *model.TuberApp, yamls *gcr.AppYamls) string {
	branch, err := gcr.TagFromRef(app.ImageTag)
	if err != nil {
		return ""
	}

	var newSHA string
	for _, tag := range yamls.Tags {
		// if you're pushing more than branch and commit sha, just.. stop that for now
		if branch != tag {
			newSHA = tag
			break
		}
	}

	if newSHA == "" {
		return ""
	}

	var oldSHA string
	for _, tag := range app.CurrentTags {
		if branch != tag {
			oldSHA = tag
			break
		}
	}

	if newSHA == "" {
		return ""
	}

	return fmt.Sprintf(" - <%s|Compare Diff>", app.GithubURL+oldSHA+"..."+newSHA)
}

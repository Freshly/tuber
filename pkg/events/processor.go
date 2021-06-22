package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/gcr"
	"github.com/freshly/tuber/pkg/report"
	"github.com/freshly/tuber/pkg/slack"
	"github.com/getsentry/sentry-go"
	"google.golang.org/api/option"

	"go.uber.org/zap"
)

// Processor processes events
type Processor struct {
	ctx                context.Context
	logger             *zap.Logger
	creds              []byte
	ClusterData        *core.ClusterData
	reviewAppsEnabled  bool
	locks              *map[string]*sync.Cond
	slackClient        *slack.Client
	db                 *core.DB
	sentryBearerToken  string
	tuberEventsProject string
	tuberEventsTopic   string
}

// NewProcessor constructs a Processor
func NewProcessor(ctx context.Context, logger *zap.Logger, db *core.DB, creds []byte, clusterData *core.ClusterData, reviewAppsEnabled bool, slackClient *slack.Client, sentryBearerToken string, tuberEventsProject string, tuberEventsTopic string) *Processor {
	l := make(map[string]*sync.Cond)

	return &Processor{
		ctx:                ctx,
		logger:             logger,
		creds:              creds,
		ClusterData:        clusterData,
		reviewAppsEnabled:  reviewAppsEnabled,
		locks:              &l,
		slackClient:        slackClient,
		db:                 db,
		sentryBearerToken:  sentryBearerToken,
		tuberEventsProject: tuberEventsProject,
		tuberEventsTopic:   tuberEventsTopic,
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
		p.slackClient.Message(event.logger, ":double_vertical_bar: release skipped for "+app.Name+" as it is paused", app.SlackChannel)
		event.logger.Warn("deployments are paused for this app; skipping", zap.String("appName", app.Name))
		cond.L.Unlock()
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
		p.slackClient.Message(logger, ":skull_and_crossbones: image or tuber layer not found for "+app.Name, app.SlackChannel)
		logger.Error("failed to find tuber layer", zap.Error(err))
		report.Error(err, errorScope.WithContext("find tuber layer"))
		return
	}

	var ti tagInfo
	if app.GithubURL != "" {
		var err error
		ti, err = getTagInfo(app, yamls)
		if err != nil {
			logger.Error("error prevented git diffs and release events for a release", zap.Error(err))
			// report.Error(err, errorScope.WithContext("error prevented git diffs and release events for a release"))
		}
	}

	startTime := time.Now()
	err = core.Release(
		p.db,
		yamls,
		logger,
		errorScope,
		app,
		event.digest,
		p.ClusterData,
		p.slackClient,
		ti.diffText,
		p.sentryBearerToken,
	)

	if err != nil {
		logger.Warn("release failed", zap.Error(err), zap.Duration("duration", time.Since(startTime)))
		p.slackClient.Message(logger, "<!here> :loudspeaker: release failed for *"+app.Name+"*", app.SlackChannel)
		return
	}

	p.slackClient.Message(logger, ":checkered_flag: *"+app.Name+"*: release complete", app.SlackChannel)
	logger.Info("release complete", zap.Duration("duration", time.Since(startTime)))

	if ti.hasEventData() {
		logger.Info("posting completed event")
		err = p.postCompleted(app, ti)
		if err != nil {
			logger.Error("error prevented sending release completed event", zap.Error(err))
			// report.Error(err, errorScope.WithContext("error prevented sending release completed event"))
		}
	}
}

type tagInfo struct {
	branch   string
	newSHA   string
	diffText string
}

func (t tagInfo) hasEventData() bool {
	return t.branch != "" && t.newSHA != ""
}

func getTagInfo(app *model.TuberApp, yamls *gcr.AppYamls) (tagInfo, error) {
	branch, err := gcr.TagFromRef(app.ImageTag)
	if err != nil {
		return tagInfo{}, err
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
		return tagInfo{}, fmt.Errorf("no git sha found in tags of incoming image")
	}

	var oldSHA string
	for _, tag := range app.CurrentTags {
		if branch != tag {
			oldSHA = tag
			break
		}
	}

	var diffText string
	if oldSHA != "" {
		diffText = fmt.Sprintf(" - <%s|Compare Diff>", "https://github.com/"+app.GithubURL+"/compare/"+oldSHA+"..."+newSHA)
	}

	return tagInfo{
		branch:   branch,
		newSHA:   newSHA,
		diffText: diffText,
	}, nil
}

type Message struct {
	AppName   string `json:"appName"`
	CommitSha string `json:"commitSha"`
	Repo      string `json:"repo"`
	Branch    string `json:"branch"`
}

func (p Processor) postCompleted(app *model.TuberApp, t tagInfo) error {
	client, err := pubsub.NewClient(p.ctx, p.tuberEventsProject, option.WithCredentialsJSON(p.creds))
	if err != nil {
		return err
	}

	topic := client.Topic(p.tuberEventsTopic)

	msg := Message{
		AppName:   app.Name,
		CommitSha: t.newSHA,
		Repo:      app.GithubURL,
		Branch:    t.branch,
	}
	marshalled, err := json.Marshal(&msg)
	if err != nil {
		return err
	}

	res := topic.Publish(p.ctx, &pubsub.Message{Data: marshalled})
	_, err = res.Get(p.ctx)
	if err != nil {
		return err
	}
	return nil
}

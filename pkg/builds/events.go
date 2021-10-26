package builds

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/gcr"
	"github.com/freshly/tuber/pkg/pubsub"
	"github.com/freshly/tuber/pkg/report"
	"github.com/freshly/tuber/pkg/slack"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

type Event struct {
	pubsub.Message
	logger     *zap.Logger
	errorScope report.Scope
}

func newEvent(logger *zap.Logger, message pubsub.Message) *Event {
	logger = logger.With(
		zap.String("repoName", message.Substitutions.RepoName),
		zap.String("branchName", message.Substitutions.BranchName),
	)

	scope := report.Scope{}

	return &Event{
		Message:    message,
		logger:     logger,
		errorScope: scope,
	}
}

func NewProcessor(ctx context.Context, logger *zap.Logger, db *core.DB, slackClient *slack.Client) *Processor {
	return &Processor{
		ctx:         ctx,
		logger:      logger,
		db:          db,
		slackClient: slackClient,
	}
}

type Processor struct {
	ctx         context.Context
	logger      *zap.Logger
	db          *core.DB
	slackClient *slack.Client
}

func (p *Processor) Process(message pubsub.Message) {
	event := newEvent(p.logger, message)

	defer sentry.Recover()

	p.notify(event)
}

func (p *Processor) notify(event *Event) {
	if event.Status != "WORKING" && event.Status != "SUCCESS" && event.Status != "FAILURE" {
		event.logger.Debug("build status received; not worth notifying", zap.String("build-status", event.Status))
		return
	}

	if event.Substitutions.BranchName == "" {
		event.logger.Debug("build notification payload missing substitutions.BRANCH_NAME")
		return
	}

	apps, err := p.appsToNotify(event, event.Substitutions.RepoName)
	if err != nil {
		event.logger.Error("failed to find apps matching repo name", zap.Error(err))
		return
	}

	for _, app := range apps {
		message := buildMessage(event, app)
		p.slackClient.Message(p.logger, message, app.SlackChannel, slack.MsgOptionDisableLinkUnfurl())
	}
}

func (p *Processor) appsToNotify(event *Event, repoName string) ([]*model.TuberApp, error) {
	var matches []*model.TuberApp
	var found []*model.TuberApp

	apps, err := p.db.AppsByCloudSourceRepo(repoName)
	if err != nil {
		return nil, err
	}
	found = append(found, apps...)

	apps, err = p.db.AppsByName(repoName)
	if err != nil {
		return nil, err
	}
	found = append(found, apps...)

	// ImageTag = Docker Ref = gcr.io/freshly-docker/appName:branchName
	for _, app := range found {
		branchName, err := gcr.TagFromRef(app.ImageTag)
		if err != nil {
			return nil, err
		}

		if branchName == event.Substitutions.BranchName {
			matches = append(matches, app)
		}
	}

	return matches, nil
}

func buildMessage(event *Event, app *model.TuberApp) string {
	var msg string
	switch event.Status {
	case "WORKING":
		msg = fmt.Sprintf(":package: Build started for *%s*:%s - <%s|Logs>", app.Name, event.Substitutions.BranchName, event.LogURL)
	case "SUCCESS":
		msg = fmt.Sprintf(":white_check_mark: Build succeeded for *%s*:%s - <%s|Logs>", app.Name, event.Substitutions.BranchName, event.LogURL)
	case "FAILURE":
		msg = fmt.Sprintf(":bomb: Build failed for *%s*:%s - <%s|Logs>", app.Name, event.Substitutions.BranchName, event.LogURL)
	}

	return msg
}

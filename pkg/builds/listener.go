package builds

import (
	"context"
	"fmt"
	"strings"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/gcr"
	"github.com/freshly/tuber/pkg/pubsub"
	"github.com/freshly/tuber/pkg/report"
	"github.com/freshly/tuber/pkg/slack"
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

func (p *Processor) ProcessMessage(message pubsub.Message) {
	event := newEvent(p.logger, message)

	if event.Status != "WORKING" && event.Status != "SUCCESS" && event.Status != "FAILED" {
		event.logger.Debug("build status received; not worth notifying", zap.String("build-status", event.Status))
		return
	}

	if event.Substitutions.BranchName == "" {
		event.logger.Debug("build notification payload missing substitutions.BRANCH_NAME")
		return
	}

	// This code is to handle Github App triggers. In those cases, event.Substitutions.RepoName is just the app name i.e. "distribution"
	// In Tuber's DB we store all cloudSoureRepos as "github_freshly_{appName}"
	repoName := event.Substitutions.RepoName
	if !strings.Contains(event.Substitutions.RepoName, "github_freshly") {
		repoName = "github_freshly_" + event.Substitutions.BranchName
	}

	matches, err := p.appsToNotify(event, repoName)
	if err != nil {
		event.logger.Error("failed to find apps matching repo name", zap.Error(err))
		return
	}

	for _, app := range matches {
		message := buildMessage(event, app)
		p.slackClient.Message(p.logger, message, app.SlackChannel)
	}
}

func (p *Processor) appsToNotify(event *Event, repoName string) ([]*model.TuberApp, error) {
	var matches []*model.TuberApp
	apps, err := p.db.AppsByCloudSourceRepo(repoName)
	if err != nil {
		return nil, err
	}

	// ImageTag = Docker Ref = gcr.io/freshly-docker/appName:branchName
	for _, app := range apps {
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
		msg = fmt.Sprintf(":building_construction: Build started for *%s* - <%s|Build Logs>", app.Name, event.LogURL)
	case "SUCCESS":
		msg = fmt.Sprintf(":white_check_mark: Build succeeded for *%s* - <%s|Build Logs>", app.Name, event.LogURL)
	case "FAILED":
		msg = fmt.Sprintf(":x: Build failed for *%s* - <%s|Build Logs>", app.Name, event.LogURL)
	}

	return msg
}

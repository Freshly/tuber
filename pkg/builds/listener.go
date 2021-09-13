package builds

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/pubsub"
	"github.com/freshly/tuber/pkg/report"
	"github.com/freshly/tuber/pkg/slack"
	"go.uber.org/zap"
)

type Event struct {
	Name       string
	Status     string
	LogURL     string
	Images     []string
	Logger     *zap.Logger
	errorScope report.Scope
}

func newEvent(logger *zap.Logger, message pubsub.Message) *Event {
	logger = logger.With()
	scope := report.Scope{}

	return &Event{
		Logger:     logger,
		errorScope: scope,
		Name:       message.Name,
		Status:     message.Status,
		Images:     message.Images,
		LogURL:     message.LogURL,
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

	if event.Status != "SUCCESS" && event.Status != "FAILED" {
		event.Logger.Debug("build status received; not worth notifying", zap.String("build-status", event.Status))
		return
	}

	if len(event.Images) < 1 {
		event.Logger.Debug("build contains no images; skipping")
		return
	}

	var apps []*model.TuberApp
	for _, img := range event.Images {
		matches, err := p.db.AppsForTag(img)
		if err != nil {
			event.Logger.Error("failed to look up tuber apps", zap.Error(err), zap.String("tag", img))
		}

		apps = append(apps, matches...)
	}

	for _, app := range apps {
		message := buildMessage(event, app)
		p.slackClient.Message(p.logger, message, app.SlackChannel)
	}
}

func buildMessage(event *Event, app *model.TuberApp) string {
	var msg string
	switch event.Status {
	case "SUCCESS":
		msg = fmt.Sprintf("Build succeeded for %s", app.Name)
	case "FAILED":
		msg = fmt.Sprintf("Build failed for %s. See logs: %s", app.Name, event.LogURL)
	}

	return msg
}

package builds

import (
	"context"
	"fmt"
	"time"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/slack"
	"go.uber.org/zap"
	"google.golang.org/api/cloudbuild/v1"
)

type Build struct {
	Status    string
	Link      string
	StartTime string
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

func (p *Processor) ProcessMessage(event *events.Event) {
	// if event.Build.Status != "FAILED" {
	// 		return
	// }

	// If we have a tag on the event, that would be SO cool
	_, err := p.db.AppsForTag(event.Tag)
	if err != nil {
		event.Logger.Error("failed to look up tuber apps", zap.Error(err))
		return
	}
}

func FindByApp(app *model.TuberApp, triggersProjectName string) ([]*model.Build, error) {
	ctx := context.Background()
	client, err := cloudbuild.NewService(ctx)
	if err != nil {
		return nil, err
	}

	buildsResponse, err := cloudbuild.NewProjectsBuildsService(client).List(triggersProjectName).PageSize(3).Filter(fmt.Sprintf(`trigger_id="%s"`, app.TriggerID)).Do()
	if err != nil {
		return nil, err
	}

	var builds []*model.Build
	for _, build := range buildsResponse.Builds {
		var startTime string
		if build.StartTime != "" {
			parsed, timeErr := time.Parse(time.RFC3339, build.StartTime)
			if timeErr == nil {
				startTime = parsed.Format(time.RFC822)
			}
		}
		builds = append(builds, &model.Build{Status: build.Status, Link: build.LogUrl, StartTime: startTime})
	}
	return builds, err
}

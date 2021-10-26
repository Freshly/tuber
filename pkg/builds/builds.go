package builds

import (
	"context"
	"fmt"
	"time"

	"github.com/freshly/tuber/graph/model"
	"google.golang.org/api/cloudbuild/v1"
)

type Build struct {
	Status    string
	Link      string
	StartTime string
}

func FindByApp(app *model.TuberApp, triggersProjectName string) ([]*model.Build, error) {
	ctx := context.Background()
	client, err := cloudbuild.NewService(ctx)
	if err != nil {
		return nil, err
	}

	if app.TriggerID == "" {
		return make([]*model.Build, 0), nil
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

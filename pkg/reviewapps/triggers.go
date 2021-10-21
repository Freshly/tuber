package reviewapps

import (
	"context"
	"fmt"

	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"
)

func CreateTrigger(service *cloudbuild.ProjectsTriggersService, repoSource cloudbuild.RepoSource, project string, reviewAppName string) (string, error) {
	buildTrigger := cloudbuild.BuildTrigger{
		Description:     "created by tuber",
		Filename:        "cloudbuild.yaml",
		Name:            reviewAppName,
		TriggerTemplate: &repoSource,
	}
	triggerCreateCall := service.Create(project, &buildTrigger)
	triggerCreateResult, err := triggerCreateCall.Do()
	if err != nil {
		return "", fmt.Errorf("create trigger: %w", err)
	}

	return triggerCreateResult.Id, nil
}

func RunTrigger(service *cloudbuild.ProjectsTriggersService, repoSource cloudbuild.RepoSource, triggerId string, project string) error {
	triggerRunCall := service.Run(project, triggerId, &repoSource)
	_, err := triggerRunCall.Do()
	if err != nil {
		return fmt.Errorf("error running trigger %s: %v", triggerId, err)
	}

	return nil
}

func deleteReviewAppTrigger(ctx context.Context, creds []byte, project string, triggerID string) error {
	cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(creds))
	if err != nil {
		return fmt.Errorf("cloudbuild service: %w", err)
	}

	service := cloudbuild.NewProjectsTriggersService(cloudbuildService)

	if _, err := service.Delete(project, triggerID).Do(); err != nil {
		return fmt.Errorf("failed to delete on cloudbuild api: %v", err)
	}

	return nil
}

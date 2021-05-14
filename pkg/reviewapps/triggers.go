package reviewapps

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"
)

// CreateAndRunTrigger creates a cloud build trigger for the review app
func CreateAndRunTrigger(ctx context.Context, logger *zap.Logger, creds []byte, project string, branch string, cloudSourceRepo string, reviewAppName string) (string, error) {
	cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(creds))
	if err != nil {
		return "", fmt.Errorf("cloudbuild service: %w", err)
	}
	service := cloudbuild.NewProjectsTriggersService(cloudbuildService)
	triggerTemplate := cloudbuild.RepoSource{
		BranchName: branch,
		ProjectId:  project,
		RepoName:   cloudSourceRepo,
	}

	buildTrigger := cloudbuild.BuildTrigger{
		Description:     "created by tuber",
		Filename:        "cloudbuild.yaml",
		Name:            reviewAppName,
		TriggerTemplate: &triggerTemplate,
	}
	triggerCreateCall := service.Create(project, &buildTrigger)
	triggerCreateResult, err := triggerCreateCall.Do()
	if err != nil {
		return "", fmt.Errorf("create trigger: %w", err)
	}

	triggerRunCall := service.Run(project, triggerCreateResult.Id, &triggerTemplate)
	_, err = triggerRunCall.Do()
	if err != nil {
		logger.Error(fmt.Sprintf("run trigger: %s", err.Error()))
		return "", fmt.Errorf("error running trigger: does the branch exist")
	}

	return triggerCreateResult.Id, nil
}

func deleteReviewAppTrigger(ctx context.Context, creds []byte, project string, triggerID string) error {
	cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(creds))
	if err != nil {
		return fmt.Errorf("cloudbuild service: %w", err)
	}

	service := cloudbuild.NewProjectsTriggersService(cloudbuildService)

	deleteCall := service.Delete(project, triggerID)
	_, err = deleteCall.Do()
	if err != nil {
		return err
	}

	return nil
}

package reviewapps

import (
	"context"
	"fmt"
	"tuber/pkg/k8s"

	"go.uber.org/zap"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"
)

const TuberReposConfig = "tuber-repos"
const TuberReviewTriggersConfig = "tuber-review-triggers"

// CreateAndRunTrigger creates a cloud build trigger for the review app
func CreateAndRunTrigger(ctx context.Context, logger *zap.Logger, creds []byte, sourceRepo string, project string, targetAppName string, branch string) error {
	config, err := k8s.GetConfigResource(TuberReposConfig, "tuber", "configmap")
	if err != nil {
		return err
	}

	var cloudSourceRepo string
	for k, v := range config.Data {
		if v == sourceRepo {
			cloudSourceRepo = k
			break
		}
	}

	if cloudSourceRepo == "" {
		return fmt.Errorf("source repo not present in tuber-repos")
	}

	cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(creds))
	if err != nil {
		return fmt.Errorf("cloudbuild service: %w", err)
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
		Name:            "review-app-for-" + targetAppName,
		TriggerTemplate: &triggerTemplate,
	}
	triggerCreateCall := service.Create(project, &buildTrigger)
	triggerCreateResult, err := triggerCreateCall.Do()
	if err != nil {
		return fmt.Errorf("create trigger: %w", err)
	}

	err = k8s.PatchConfigMap(TuberReviewTriggersConfig, "tuber", targetAppName, triggerCreateResult.Id)
	if err != nil {
		return err
	}

	triggerRunCall := service.Run(project, triggerCreateResult.Id, &triggerTemplate)
	_, err = triggerRunCall.Do()
	if err != nil {
		logger.Error(fmt.Sprintf("run trigger: %s", err.Error()))
		return fmt.Errorf("error running trigger: does the branch exist")
	}

	return nil
}

func deleteReviewAppTrigger(ctx context.Context, creds []byte, project string, reviewAppName string) error {
	config, err := k8s.GetConfigResource(TuberReviewTriggersConfig, "tuber", "configmap")
	if err != nil {
		return err
	}

	triggerID := config.Data[reviewAppName]

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

	return k8s.RemoveConfigMapEntry(TuberReviewTriggersConfig, "tuber", reviewAppName)
}

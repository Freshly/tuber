package reviewapps

import (
	"context"
	"fmt"
	"tuber/pkg/k8s"

	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"
)

const tuberReposConfig = "tuber-repos"
const tuberReviewTriggersConfig = "tuber-review-triggers"

// CreateAndRunTrigger creates a cloud build trigger for the review app
func CreateAndRunTrigger(ctx context.Context, creds []byte, sourceRepo string, project string, targetAppName string, branch string) (func() error, error) {
	config, err := k8s.GetConfigResource(tuberReposConfig, "tuber", "configmap")
	if err != nil {
		return nil, err
	}

	var cloudSourceRepo string
	for k, v := range config.Data {
		if v == sourceRepo {
			cloudSourceRepo = k
			break
		}
	}

	if cloudSourceRepo == "" {
		return nil, fmt.Errorf("source repo not present in tuber-repos")
	}

	cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(creds))
	if err != nil {
		return nil, fmt.Errorf("cloudbuild service: %w", err)
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
		return nil, fmt.Errorf("create trigger: %w", err)
	}

	delete := func() error { return deleteTrigger(service, project, targetAppName, triggerCreateResult.Id) }

	err = k8s.PatchConfigMap(tuberReviewTriggersConfig, "tuber", targetAppName, triggerCreateResult.Id)
	if err != nil {
		deleteErr := deleteReviewAppTrigger(service, project, triggerCreateResult.Id, targetAppName)
		if deleteErr != nil {
			return delete, fmt.Errorf(err.Error() + deleteErr.Error())
		}
		return delete, err
	}

	triggerRunCall := service.Run(project, triggerCreateResult.Id, &triggerTemplate)
	_, err = triggerRunCall.Do()
	if err != nil {
		deleteErr := deleteReviewAppTrigger(service, project, triggerCreateResult.Id, targetAppName)
		if deleteErr != nil {
			return delete, fmt.Errorf("delete trigger: %v - %v", err, deleteErr)
		}
		return delete, fmt.Errorf("run trigger: %w", err)
	}

	return delete, nil
}

func deleteTrigger(service *cloudbuild.ProjectsTriggersService, projectID, triggerID, appName string) error {
	err := deleteReviewAppTrigger(service, projectID, triggerID, appName)
	if err != nil {
		return err
	}

	return nil
}

func deleteReviewAppTrigger(service *cloudbuild.ProjectsTriggersService, project string, triggerID string, appName string) error {
	deleteCall := service.Delete(project, triggerID)
	_, err := deleteCall.Do()
	if err != nil {
		return err
	}
	return k8s.RemoveConfigMapEntry(tuberReviewTriggersConfig, "tuber", appName)
}

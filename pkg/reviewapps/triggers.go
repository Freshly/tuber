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

func CreateAndRunTrigger(ctx context.Context, creds []byte, sourceRepo string, project string, targetAppName string, branch string) error {
	config, err := k8s.GetConfigResource(tuberReposConfig, "tuber", "configmap")
	if err != nil {
		return err
	}
	var cloudSourceRepo string
	for k, v := range config.Data {
		if v == sourceRepo {
			cloudSourceRepo = k
		}
	}
	if cloudSourceRepo == "" {
		return fmt.Errorf("source repo not present in tuber-repos")
	}
	cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(creds))
	if err != nil {
		return err
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
		return err
	}
	err = k8s.PatchConfigMap(tuberReviewTriggersConfig, "tuber", targetAppName, triggerCreateResult.Id)
	if err != nil {
		deleteErr := deleteReviewAppTrigger(service, project, triggerCreateResult.Id, targetAppName)
		if deleteErr != nil {
			return fmt.Errorf(err.Error() + deleteErr.Error())
		}
		return err
	}
	triggerRunCall := service.Run(project, triggerCreateResult.Id, &triggerTemplate)
	_, err = triggerRunCall.Do()
	if err != nil {
		deleteErr := deleteReviewAppTrigger(service, project, triggerCreateResult.Id, targetAppName)
		if deleteErr != nil {
			return fmt.Errorf(err.Error() + deleteErr.Error())
		}
		return err
	}
	return nil
}

func deleteReviewAppTrigger(service *cloudbuild.ProjectsTriggersService, project string, triggerId string, appName string) error {
	deleteCall := service.Delete(project, triggerId)
	_, err := deleteCall.Do()
	if err != nil {
		return err
	}
	return k8s.RemoveConfigMapEntry(tuberReviewTriggersConfig, "tuber", appName)
}

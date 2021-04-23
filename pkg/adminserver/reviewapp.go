package adminserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/reviewapps"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/cloudbuild/v1"
)

type Build struct {
	Status    string
	Link      string
	StartTime string
}

type reviewAppResponse struct {
	Title                 string
	Error                 string
	Name                  string
	Link                  string
	SourceAppName         string
	SuccessfulBuildExists bool
	Builds                []Build
}

func (s server) reviewApp(c *gin.Context) {
	template := "reviewApp.html"
	reviewAppName := c.Param("reviewAppName")

	data := reviewAppResponse{
		Title:         fmt.Sprintf("Tuber Admin: %s", reviewAppName),
		SourceAppName: c.Param("appName"),
		Name:          reviewAppName,
	}

	if !s.reviewAppsEnabled {
		data.Error = "review apps are not enabled on this cluster"
		c.HTML(http.StatusNotFound, template, data)
		return
	}

	data.Link = fmt.Sprintf("https://%s.%s/", reviewAppName, s.clusterDefaultHost)

	builds, err := reviewAppBuilds(reviewAppName, s.triggersProjectName, s.cloudbuildClient)
	if err != nil {
		s.logger.Error(fmt.Sprintf("error pulling review app builds for %s: %v", reviewAppName, err))
		data.Error = "error retrieving review app builds, try refreshing"
		c.HTML(http.StatusInternalServerError, template, data)
		return
	}

	data.Builds = builds

	for _, build := range data.Builds {
		if build.Status == "SUCCESS" {
			data.SuccessfulBuildExists = true
			break
		}
	}

	c.HTML(http.StatusOK, template, data)
}

func reviewAppBuilds(reviewAppName string, triggersProjectName string, cloudbuildClient *cloudbuild.Service) ([]Build, error) {
	config, err := k8s.GetConfigResource(reviewapps.TuberReviewTriggersConfig, "tuber", "configmap")
	if err != nil {
		return nil, fmt.Errorf("review triggers map not found")
	}

	triggerId := config.Data[reviewAppName]
	if triggerId == "" {
		return nil, fmt.Errorf("trigger is untracked or it doesnt exist")
	}

	buildsResponse, err := cloudbuild.NewProjectsBuildsService(cloudbuildClient).List(triggersProjectName).Filter(fmt.Sprintf(`trigger_id="%s"`, triggerId)).Do()
	if err != nil {
		return nil, err
	}

	var builds []Build
	for _, build := range buildsResponse.Builds {
		var startTime string
		if build.StartTime != "" {
			parsed, timeErr := time.Parse(time.RFC3339, build.StartTime)
			if timeErr == nil {
				startTime = parsed.Format(time.RFC822)
			}
		}
		builds = append(builds, Build{Status: build.Status, Link: build.LogUrl, StartTime: startTime})
	}
	return builds, err
}

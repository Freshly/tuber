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
	Title             string
	Error             string
	Name              string
	Link              string
	SourceAppName     string
	NoSuccessfulBuild bool
	Builds            []Build
}

func (s server) reviewApp(c *gin.Context) {
	template := "reviewApp.html"
	sourceAppName := c.Param("appName")
	reviewAppName := c.Param("reviewAppName")
	var status = http.StatusOK

	data := reviewAppResponse{
		Title:         fmt.Sprintf("Tuber Admin: %s", reviewAppName),
		SourceAppName: sourceAppName,
		Name:          reviewAppName,
	}

	if !s.reviewAppsEnabled {
		data.Error = "review apps are not enabled on this cluster silly"
		c.HTML(http.StatusNotFound, template, data)
		return
	}

	data.Link = fmt.Sprintf("https://%s.%s/", reviewAppName, s.clusterDefaultHost)

	builds, err := reviewAppBuilds(reviewAppName, s.triggersProjectName, s.cloudbuildClient)
	if err != nil {
		data.Error = err.Error()
		c.HTML(http.StatusInternalServerError, template, data)
		return
	}

	var successfulBuildExists bool
	for _, build := range builds {
		if build.Status == "SUCCESS" {
			successfulBuildExists = true
			break
		}
	}
	data.NoSuccessfulBuild = !successfulBuildExists

	data.Builds = builds
	c.HTML(status, template, data)
}

func reviewAppBuilds(reviewAppName string, triggersProjectName string, cloudbuildClient *cloudbuild.Service) ([]Build, error) {
	config, err := k8s.GetConfigResource(reviewapps.TuberReviewTriggersConfig, "tuber", "configmap")
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

package adminserver

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/freshly/tuber/pkg/core"
	"github.com/gin-gonic/gin"
)

type appResponse struct {
	Title                  string
	Error                  string
	ReviewAppsEnabled      bool
	Name                   string
	ReviewAppCreationError string
	ReviewApps             []appReviewApp
	Link                   string
}

type appReviewApp struct {
	Name   string
	Branch string
}

func (s server) app(c *gin.Context) {
	template := "app.html"
	appName := c.Param("appName")

	response := &appResponse{
		Title:                  fmt.Sprintf("Tuber Admin: %s", appName),
		ReviewAppsEnabled:      s.reviewAppsEnabled,
		Name:                   appName,
		ReviewAppCreationError: c.Query("error"),
	}

	if s.reviewAppsEnabled {
		reviewApps, err := reviewApps(appName)
		if err != nil {
			response.Error = err.Error()
			c.HTML(http.StatusInternalServerError, template, response)
			return
		}
		response.ReviewApps = reviewApps
	}

	response.Link = fmt.Sprintf("https://%s.%s/", appName, s.clusterDefaultHost)
	c.HTML(http.StatusOK, template, response)
}

func reviewApps(sourceAppName string) ([]appReviewApp, error) {
	allReviewApps, err := core.TuberReviewApps()
	if err != nil {
		return nil, err
	}

	sourceApps, err := core.TuberSourceApps()
	if err != nil {
		return nil, err
	}

	sourceApp, err := sourceApps.FindApp(sourceAppName)
	if err != nil {
		return nil, err
	}

	var reviewAppsList core.AppList
	for _, reviewApp := range allReviewApps {
		if sourceApp.Repo == reviewApp.Repo {
			reviewAppsList = append(reviewAppsList, reviewApp)
		}
	}

	sort.Slice(reviewAppsList, func(i, j int) bool {
		return reviewAppsList[i].Name < reviewAppsList[j].Name
	})

	var reviewApps []appReviewApp
	for _, app := range reviewAppsList {
		reviewApps = append(reviewApps, appReviewApp{Name: app.Name, Branch: app.Tag})
	}

	return reviewApps, err
}

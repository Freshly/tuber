package adminserver

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/db"
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
	Name     string
	ImageTag string
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
		reviewApps, err := reviewApps(appName, s.db)
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

func reviewApps(sourceAppName string, d *db.DB) ([]appReviewApp, error) {
	var app *model.TuberApp
	err := d.Find("apps", sourceAppName, app)
	if err != nil {
		return nil, err
	}

	var reviewAppsList []*model.TuberApp
	if err := d.Get("apps", reviewAppsList, db.Q().String("sourceAppName", app.Name).Bool("reviewApp", true)); err != nil {
		return nil, err
	}

	sort.Slice(reviewAppsList, func(i, j int) bool {
		return reviewAppsList[i].Name < reviewAppsList[j].Name
	})

	var reviewApps []appReviewApp
	for _, app := range reviewAppsList {
		reviewApps = append(reviewApps, appReviewApp{Name: app.Name, ImageTag: app.ImageTag})
	}

	return reviewApps, err
}

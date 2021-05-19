package adminserver

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/freshly/tuber/pkg/core"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type dashboardResponse struct {
	Title      string
	Error      string
	SourceApps []sourceApp
}

type sourceApp struct {
	Name   string
	Branch string
}

func (s server) dashboard(c *gin.Context) {
	sourceApps, err := sourceApps(s.logger, s.db)
	var status = http.StatusOK
	data := dashboardResponse{Title: "Tuber Dashboard"}

	if err != nil {
		status = http.StatusInternalServerError
		s.logger.Error("error rendering dashboard: " + err.Error())
		data.Error = "internal error pulling source apps for dashboard"
	} else {
		data.SourceApps = sourceApps
	}

	c.HTML(status, "dashboard.html", data)
}

func sourceApps(logger *zap.Logger, db *core.DB) ([]sourceApp, error) {
	tuberApps, err := db.SourceApps()
	if err != nil {
		logger.Error("error retrieving source apps", zap.Error(err))
		return []sourceApp{}, fmt.Errorf("error retrieving source apps")
	}

	sort.Slice(tuberApps, func(i, j int) bool {
		return tuberApps[i].Name < tuberApps[j].Name
	})

	var apps []sourceApp

	for _, app := range tuberApps {
		apps = append(apps, sourceApp{Name: app.Name, Branch: app.ImageTag})
	}

	return apps, nil
}

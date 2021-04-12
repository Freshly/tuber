package adminserver

import (
	"net/http"

	"github.com/freshly/tuber/pkg/reviewapps"
	"github.com/gin-gonic/gin"
)

func (s server) createReviewApp(c *gin.Context) {
	branch := c.PostForm("branch")
	appName := c.Param("appName")
	reviewAppName, err := reviewapps.CreateReviewApp(s.ctx, s.logger, branch, appName, s.creds, s.triggersProjectName)
	if err == nil {
		c.Redirect(http.StatusFound, "reviewapps/"+reviewAppName)
	} else {
		errString := "review app creation error: " + err.Error()
		c.Redirect(http.StatusFound, "?error="+errString)
	}
}

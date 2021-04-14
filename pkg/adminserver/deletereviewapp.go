package adminserver

import (
	"net/http"

	"github.com/freshly/tuber/pkg/reviewapps"
	"github.com/gin-gonic/gin"
)

func (s server) deleteReviewApp(c *gin.Context) {
	sourceAppName := c.Param("appName")
	reviewAppName := c.Param("reviewAppName")
	err := reviewapps.DeleteReviewApp(s.ctx, reviewAppName, s.creds, s.triggersProjectName)
	if err == nil {
		c.Redirect(http.StatusFound, "/tuber/apps/"+sourceAppName)
	} else {
		errString := "review app deletion error: " + err.Error()
		c.Redirect(http.StatusFound, "/tuber/apps/"+sourceAppName+"?error="+errString)
	}
}

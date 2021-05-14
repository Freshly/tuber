package adminserver

import (
	"net/http"

	"github.com/freshly/tuber/pkg/reviewapps"
	"github.com/gin-gonic/gin"
)

func (s server) createReviewApp(c *gin.Context) {
	reviewAppName, err := reviewapps.CreateReviewApp(c.Request.Context(), s.db, s.logger, c.PostForm("branch"), c.Param("appName"), s.creds, s.triggersProjectName)
	if err == nil {
		c.Redirect(http.StatusFound, "reviewapps/"+reviewAppName)
	} else {
		s.logger.Error("review app creation error: " + err.Error())
		clientError := "review app creation failed"
		c.Redirect(http.StatusFound, "?error="+clientError)
	}
}

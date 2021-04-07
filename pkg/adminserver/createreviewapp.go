package adminserver

import (
	"net/http"

	"github.com/freshly/tuber/pkg/reviewapps"
)

func createReviewAppHandler() createReviewApp {
	return createReviewApp{
		errorParam:   "error",
		creationPath: "createReviewApp",
	}
}

type createReviewApp struct {
	reviewAppsEnabled     bool
	creationArgs          *createReviewAppsArgs
	triggersProjectName   string
	reviewAppCreatedParam string
	errorParam            string
	creationPath          string
}

func (c createReviewApp) path() string { return sourceAppHandler().path() + "/" + c.creationPath }

func (c createReviewApp) setup(reviewAppsEnabled bool, creationArgs *createReviewAppsArgs, triggersProjectName string) handler {
	c.reviewAppsEnabled = reviewAppsEnabled
	c.creationArgs = creationArgs
	c.triggersProjectName = triggersProjectName
	return c
}

func (c createReviewApp) handle(w http.ResponseWriter, r *http.Request) {
	branch := r.FormValue("branch")
	appName := r.FormValue("appname")
	reviewAppName, err := reviewapps.CreateReviewApp(c.creationArgs.ctx, c.creationArgs.logger, branch, appName, c.creationArgs.creds, c.triggersProjectName)
	if err == nil {
		http.Redirect(w, r, reviewAppHandler().pathPrefix+"/"+reviewAppName, http.StatusSeeOther)
	} else {
		errString := "review app creation error: " + err.Error()
		http.Redirect(w, r, "?"+c.errorParam+"="+errString, http.StatusSeeOther)
	}
}

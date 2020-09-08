package reviewapps

import (
	"tuber/pkg/core"
	"tuber/pkg/k8s"

	"github.com/spf13/viper"
)

func canCreate(appName, token string) bool {
	return appName != "tuber" &&
		viper.GetBool("review-apps") &&
		k8s.CanDeploy(appName, token) &&
		appExists(appName)
}

func appExists(appName string) bool {
	apps, err := core.SourceAndReviewApps()
	if err != nil {
		return false
	}

	for _, app := range apps {
		if app.Name == appName {
			return true
		}
	}

	return false
}

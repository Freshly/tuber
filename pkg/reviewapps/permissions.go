package reviewapps

import (
	"tuber/pkg/core"
	"tuber/pkg/k8s"

	"go.uber.org/zap"
)

func canCreate(logger *zap.Logger, appName, token string) (bool, error) {
	if appName == "tuber" {
		return false, nil
	}

	exists, err := appExists(appName)
	if err != nil || !exists {
		return false, err
	}

	return k8s.CanDeploy(appName, token), nil
}

func appExists(appName string) (bool, error) {
	apps, err := core.SourceAndReviewApps()
	if err != nil {
		return false, err
	}

	for _, app := range apps {
		if app.Name == appName {
			return true, nil
		}
	}

	return false, nil
}

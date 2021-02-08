package reviewapps

import (
	"tuber/pkg/core"
	"tuber/pkg/k8s"

	"go.uber.org/zap"
)

func canCreate(logger *zap.Logger, appName, token string) (bool, error) {
	exists, err := appExists(appName)
	if err != nil {
		return false, err
	}

	canDeploy := k8s.CanDeploy(appName, token)

	return appName != "tuber" && canDeploy && exists, nil
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

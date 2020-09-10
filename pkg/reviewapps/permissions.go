package reviewapps

import (
	"tuber/pkg/core"
	"tuber/pkg/k8s"

	"go.uber.org/zap"
)

func canCreate(logger *zap.Logger, appName, token string) bool {
	logger.Info("----------- canCreate --------------")
	logger.Info("canDeploy ->", zap.Bool("canDeploy", k8s.CanDeploy(appName, token)))
	logger.Info("appExists ->", zap.Bool("appExists", appExists(appName)))

	return appName != "tuber" &&
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

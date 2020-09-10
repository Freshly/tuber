package reviewapps

import (
	"fmt"
	"tuber/pkg/core"
	"tuber/pkg/k8s"
)

func canCreate(appName, token string) bool {
	fmt.Println("----------- canCreate --------------")
	fmt.Println("appName ->", appName)
	fmt.Println("canDeploy ->", k8s.CanDeploy(appName, token))
	fmt.Println("appExists ->", appExists(appName))

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

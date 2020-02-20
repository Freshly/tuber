package core

import "tuber/pkg/k8s"

// DestroyTuberApp deletes all resources for the given app on the current cluster
func DestroyTuberApp(appName string) (err error) {
	if err = k8s.Delete("namespace", appName, appName); err != nil {
		return
	}

	if err = RemoveAppConfig(appName); err != nil {
		return
	}

	return
}

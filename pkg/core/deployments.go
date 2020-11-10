package core

import "tuber/pkg/k8s"

// ReleasesPaused checks the tuber-app-pauses configmap for the presence of the given app
func ReleasesPaused(appName string) (bool, error) {
	config, err := k8s.GetConfigResource("tuber-app-pauses", "tuber", "ConfigMap")
	if err != nil {
		return false, err
	}

	return config.Data[appName] == "true", nil
}

// PauseDeployments adds an app to the tuber-app-pauses configmap
func PauseDeployments(appName string) error {
	exists, err := k8s.Exists("configmap", "tuber-app-pauses", "tuber")
	if err != nil {
		return err
	}

	if !exists {
		if err = k8s.Create("tuber", "configmap", "tuber-app-pauses"); err != nil {
			return err
		}
	}

	return k8s.PatchConfigMap("tuber-app-pauses", "tuber", appName, "true")
}

// ResumeDeployments removes an app from the tuber-app-pauses configmap
func ResumeDeployments(appName string) error {
	exists, err := k8s.Exists("configmap", "tuber-app-pauses", "tuber")
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	return k8s.RemoveConfigMapEntry("tuber-app-pauses", "tuber", appName)
}

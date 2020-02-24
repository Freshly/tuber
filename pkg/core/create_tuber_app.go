package core

import (
	"fmt"
	data "tuber/data/tuberapps"
	"tuber/pkg/k8s"
)

// CreateTuberApp adds a new tuber app configuration, including namespace,
// role, rolebinding, and a listing in tuber-apps
func CreateTuberApp(appName string, repo string, tag string) error {
	namespaceData := map[string]string{
		"namespace": appName,
	}

	existsAlready, err := k8s.Exists("namespace", appName, appName)
	if err != nil {
		return err
	}

	if existsAlready {
		return AddAppConfig(appName, repo, tag)
	} else {
		err := newAppSetup(appName, namespaceData)
		if err != nil {
			return err
		}
		return AddAppConfig(appName, repo, tag)
	}
}

func newAppSetup(appName string, namespaceData map[string]string) error {
	var err error
	for _, yaml := range []data.TuberYaml{data.Namespace, data.Role, data.Rolebinding} {
		err = ApplyTemplate(appName, string(yaml.Contents), namespaceData)
	}
	if err == nil {
		err = k8s.CreateEnv(appName)
	}

	if err != nil {
		deleteErr := k8s.Delete("namespace", appName, appName)
		if deleteErr != nil {
			return fmt.Errorf(err.Error() + " failed delete: " + deleteErr.Error())
		}
		return err
	}

	return nil
}

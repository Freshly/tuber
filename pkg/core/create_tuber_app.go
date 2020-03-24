package core

import (
	"fmt"
	yamls "tuber/data/tuberapps"
	"tuber/pkg/k8s"
)

// NewAppSetup adds a new tuber app configuration, including namespace,
// role, rolebinding, and a listing in tuber-apps
func NewAppSetup(appName string, istio bool) error {
	var err error
	var istioEnabled string
	if istio {
		istioEnabled = "enabled"
	} else {
		istioEnabled = "disabled"
	}

	data := map[string]string{
		"namespace":    appName,
		"istioEnabled": istioEnabled,
	}

	for _, yaml := range []yamls.TuberYaml{yamls.Namespace, yamls.Role, yamls.Rolebinding} {
		err = ApplyTemplate(appName, string(yaml.Contents), data)
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

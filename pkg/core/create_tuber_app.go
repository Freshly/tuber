package core

import (
	"github.com/freshly/tuber/pkg/k8s"

	yamls "github.com/freshly/tuber/data/tuberapps"
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
		if err != nil {
			return err
		}
	}

	existsAlready, err := k8s.Exists("secret", appName+"-env", appName)
	if err != nil {
		return err
	}

	if !existsAlready {
		err = k8s.CreateEnv(appName)
	}

	if err != nil {
		return err
	}

	return nil
}

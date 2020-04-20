package core

import (
	yamls "tuber/data/tuberapps"
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

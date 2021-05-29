package core

import (
	"fmt"

	yamls "github.com/freshly/tuber/data/tuberapps"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"
)

// DestroyTuberApp deletes all resources for the given app on the current cluster
func DestroyTuberApp(db *DB, app *model.TuberApp) error {
	if err := k8s.Delete("namespace", app.Name, app.Name); err != nil {
		return fmt.Errorf("k8s.Delete failed: %v", err)
	}

	if err := db.DeleteApp(app); err != nil {
		return fmt.Errorf("db.DeleteApp failed: %v", err)
	}

	return nil
}

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

package core

import (
	"fmt"
	"strings"

	yamls "github.com/freshly/tuber/data/tuberapps"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/containers"
	"github.com/freshly/tuber/pkg/k8s"
)

// GetRepositoryLocation returns a RepositoryLocation struct for a given Tuber App
func GetRepositoryLocation(ta *model.TuberApp) (containers.RepositoryLocation, error) {
	split := strings.SplitN(ta.ImageTag, ":", 2)
	if len(split) != 2 {
		return containers.RepositoryLocation{}, fmt.Errorf("app image tag invalid")
	}

	repoSplit := strings.SplitN(split[0], "/", 2)
	if len(repoSplit) != 2 {
		return containers.RepositoryLocation{}, fmt.Errorf("app image tag invalid")
	}

	return containers.RepositoryLocation{
		Host: repoSplit[0],
		Path: repoSplit[1],
		Tag:  split[1],
	}, nil
}

func RepoFromTag(tag string) (string, error) {
	split := strings.SplitN(tag, ":", 2)
	if len(split) != 2 {
		return "", fmt.Errorf("app image tag invalid")
	}
	return split[0], nil
}

// DestroyTuberApp deletes all resources for the given app on the current cluster
func DestroyTuberApp(db *Data, app *model.TuberApp) error {
	err := k8s.Delete("namespace", app.Name, app.Name)
	if err != nil {
		return err
	}
	err = db.DeleteTuberApp(app)
	if err != nil {
		return err
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

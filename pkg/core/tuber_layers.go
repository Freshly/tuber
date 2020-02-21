package core

import (
	"strings"
	"tuber/pkg/k8s"
)

// ReleaseTubers combines and interpolates with tuber's conventions, and applies them
func ReleaseTubers(tubers []string, app *TuberApp, digest string, data *ClusterData) error {
	interpolatables, err := tuberData(digest, app, data)
	if err != nil {
		return err
	}
	return ApplyTemplate(app.Name, strings.Join(tubers, "---\n"), interpolatables)
}

// ClusterData is configurable, cluster-wide data available for yaml interpolation
type ClusterData struct {
	DefaultGateway string
	DefaultHost    string
}

func tuberData(digest string, app *TuberApp, clusterData *ClusterData) (map[string]string, error) {
	generalData := map[string]string{
		"tuberImage":            digest,
		"clusterDefaultGateway": clusterData.DefaultGateway,
		"clusterDefaultHost":    clusterData.DefaultHost,
		"tuberAppName":          app.Name,
	}
	valuesMapExists, err := k8s.Exists("configmap", app.Name+"-values", app.Name)
	if err != nil {
		return nil, err
	}
	if valuesMapExists {
		return joinInterpolatables(generalData, app.Name)
	} else {
		return generalData, nil
	}
}

func joinInterpolatables(data map[string]string, name string) (map[string]string, error) {
	config, err := k8s.GetConfig(name+"-values", name, "configmap")
	if err != nil {
		return nil, err
	}
	for k, v := range config.Data {
		data[k] = v
	}
	return data, nil
}

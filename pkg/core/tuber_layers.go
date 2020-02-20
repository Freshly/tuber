package core

import (
	"strings"
)

// ReleaseTubers combines and interpolates with tuber's conventions, and applies them
func ReleaseTubers(tubers []string, app *TuberApp, digest string, data *ClusterData) error {
	return ApplyTemplate(app.Name, strings.Join(tubers, "---\n"), tuberData(digest, app, data))
}

// ClusterData is configurable, cluster-wide data available for yaml interpolation
type ClusterData struct {
	DefaultGateway string
	DefaultHost    string
}

func tuberData(digest string, app *TuberApp, clusterData *ClusterData) (data map[string]string) {
	return map[string]string{
		"tuberImage":            digest,
		"clusterDefaultGateway": clusterData.DefaultGateway,
		"clusterDefaultHost":    clusterData.DefaultHost,
		"tuberAppName":          app.Name,
	}
}

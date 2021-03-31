package core

import (
	"bytes"
	"text/template"
	"tuber/pkg/k8s"
)

// ApplyTemplate interpolates and applies a yaml to a given namespace
func ApplyTemplate(namespace string, templateString string, data map[string]string) error {
	interpolated, err := interpolate(templateString, data)
	if err != nil {
		return err
	}
	return k8s.Apply(interpolated, namespace)
}

func interpolate(templateString string, data map[string]string) (interpolated []byte, err error) {
	tpl, err := template.New("").Parse(templateString)

	if err != nil {
		return
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)

	if err != nil {
		return
	}

	interpolated = buf.Bytes()
	return
}

// ClusterData is configurable, cluster-wide data available for yaml interpolation
type ClusterData struct {
	DefaultGateway string
	DefaultHost    string
}

func releaseData(digest string, app *TuberApp, clusterData *ClusterData) (data map[string]string) {
	return map[string]string{
		"tuberImage":            digest,
		"clusterDefaultGateway": clusterData.DefaultGateway,
		"clusterDefaultHost":    clusterData.DefaultHost,
		"tuberAppName":          app.Name,
	}
}

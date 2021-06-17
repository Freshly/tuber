package core

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"
)

// ApplyTemplate interpolates and applies a yaml to a given namespace
func ApplyTemplate(namespace string, templateString string, data map[string]string) error {
	interpolated, err := interpolate(templateString, data)
	if err != nil {
		return err
	}
	return k8s.Apply(interpolated, namespace)
}

// BypassReleaser is for when you're feeling frisky and want to cowboy code
func BypassReleaser(app *model.TuberApp, imageTagWithDigest string, yamls []string, data *ClusterData) error {
	var interpolated [][]byte
	for _, y := range yamls {
		i, err := interpolate(y, releaseData(imageTagWithDigest, app, data))
		if err != nil {
			return fmt.Errorf("interpolation error prior to apply: %v", err)
		}
		interpolated = append(interpolated, i)
	}

	var errors []error
	for _, resource := range interpolated {
		applyErr := k8s.Apply(resource, app.Name)
		if applyErr != nil {
			errors = append(errors, applyErr)
			continue
		}
	}

	if len(errors) != 0 {
		combined := "partial apply performed, errors applying resources: "
		for _, e := range errors {
			combined = combined + e.Error() + ", "
		}
		return fmt.Errorf(strings.TrimSuffix(combined, ", "))
	}

	return nil
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
	AdminGateway   string
	AdminHost      string
}

func releaseData(digest string, app *model.TuberApp, clusterData *ClusterData) (data map[string]string) {
	vars := map[string]string{
		"tuberImage":            digest,
		"clusterDefaultGateway": clusterData.DefaultGateway,
		"clusterDefaultHost":    clusterData.DefaultHost,
		"clusterAdminGateway":   clusterData.AdminGateway,
		"clusterAdminHost":      clusterData.AdminHost,
		"tuberAppName":          app.Name,
	}

	if app.Vars != nil {
		for _, tuple := range app.Vars {
			vars[tuple.Key] = tuple.Value
		}
	}

	return vars
}

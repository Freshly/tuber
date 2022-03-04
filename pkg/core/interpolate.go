package core

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/goccy/go-yaml"
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
	interplationData := releaseData(imageTagWithDigest, app, data)
	for _, y := range yamls {
		i, err := interpolate(y, interplationData)
		if err != nil {
			return fmt.Errorf("interpolation error prior to apply: %v", err)
		}

		split := strings.Split(string(i), "\n---\n")
		for _, s := range split {
			interpolated = append(interpolated, []byte(s))
		}
		interpolated = append(interpolated, i)
	}

	exc, err := exclusions(app, interplationData)
	if err != nil {
		return fmt.Errorf("error pulling  prior to apply: %v", err)
	}

	var resources appResources

	for _, resourceYaml := range interpolated {
		var parsed parsedResource
		err = yaml.Unmarshal(resourceYaml, &parsed)
		if err != nil {
			return fmt.Errorf("unmarshal error prior to apply: %v", err)
		}

		if exc[exclusionKey(parsed.Kind, parsed.Metadata.Name)] {
			continue
		}

		resource := appResource{
			kind:     parsed.Kind,
			name:     parsed.Metadata.Name,
			contents: resourceYaml,
		}

		if resource.canBeManaged() {
			resources = append(resources, resource)
		}
	}

	var errors []error
	for _, resource := range resources {
		applyErr := k8s.Apply(resource.contents, app.Name)
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
	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal(interpolated, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", m)

	d, err := yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return d, err
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

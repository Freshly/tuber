package core

import (
	"fmt"
	"strings"
	"tuber/pkg/k8s"

	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// ReleaseTubers combines and interpolates with tuber's conventions, and applies them
func ReleaseTubers(tubers []string, app *TuberApp, digest string, data *ClusterData) ([]string, error) {
	var releaseIDs []string
	var interpolated []string
	d, err := tuberData(digest, app, data)
	if err != nil {
		return []string{}, err
	}

	for _, tuber := range tubers {
		i, interpolateErr := interpolate(tuber, d)
		interpolated = append(interpolated, string(i))
		if err != nil {
			return []string{}, interpolateErr
		}
	}

	labeled, releaseIDs, err := annotateDeployments(interpolated)
	if err != nil {
		return []string{}, err
	}

	releaseYamls := strings.Join(labeled, "---\n")

	err = k8s.Apply([]byte(releaseYamls), app.Name)
	if err != nil {
		return []string{}, err
	}
	return releaseIDs, nil
}

// ClusterData is configurable, cluster-wide data available for yaml interpolation
type ClusterData struct {
	DefaultGateway string
	DefaultHost    string
}

func tuberData(digest string, app *TuberApp, clusterData *ClusterData) (map[string]string, error) {
	universalData := map[string]string{
		"tuberImage":            digest,
		"clusterDefaultGateway": clusterData.DefaultGateway,
		"clusterDefaultHost":    clusterData.DefaultHost,
		"tuberAppName":          app.Name,
	}
	valuesMapExists, err := k8s.Exists("configmap", "tuber-config", app.Name)
	if err != nil {
		return nil, err
	}

	// Will soon be required, must always exist soon.
	// Once we're there, skip the exists check and remove this condition
	if valuesMapExists {
		return withAppSpecificData(universalData, app.Name)
	} else {
		return universalData, nil
	}
}

func withAppSpecificData(data map[string]string, name string) (map[string]string, error) {
	config, err := k8s.GetConfig("tuber-config", name, "ConfigMap")
	if err != nil {
		return nil, err
	}
	for k, v := range config.Data {
		data[k] = v
	}
	return data, nil
}

func annotateDeployments(t []string) ([]string, []string, error) {
	var tubers []string
	var releaseIDs []string
	for _, tuber := range t {
		var resource k8sResource
		b := []byte(tuber)
		err := yaml.Unmarshal(b, &resource)
		if err != nil {
			return []string{}, []string{}, err
		}
		if resource.Kind == "Deployment" {
			annotated, releaseID, err := addAnnotation(b)
			if err != nil {
				return []string{}, []string{}, err
			}
			releaseIDs = append(releaseIDs, releaseID)
			tubers = append(tubers, annotated)
		} else {
			tubers = append(tubers, tuber)
		}
	}
	return tubers, releaseIDs, nil
}

type k8sResource struct {
	ApiVersion string
	Kind       string
}

func addAnnotation(resource []byte) (string, string, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, versionKind, err := decode(resource, nil, nil)
	if err != nil {
		return "", "", err
	}
	if versionKind.Version != "v1" {
		return "", "", fmt.Errorf("must use v1 deployments")
	}

	deployment := obj.(*v1.Deployment)
	annotations := deployment.Spec.Template.ObjectMeta.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	releaseID := uuid.New().String()
	annotations["tuber/releaseID"] = releaseID
	deployment.Spec.Template.ObjectMeta.SetAnnotations(annotations)

	annotated, err := yaml.Marshal(deployment)
	if err != nil {
		return "", "", err
	}
	return string(annotated), releaseID, nil
}

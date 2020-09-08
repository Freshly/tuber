package reviewapps

import (
	"encoding/json"
	"fmt"
	"strings"
	"tuber/pkg/core"
	"tuber/pkg/k8s"
)

// NewReviewAppSetup replicates a namespace and its roles, rolebindings, and opaque secrets after removing their non-generic metadata.
// Also renames source app name matches across all relevant resources.
func NewReviewAppSetup(sourceApp string, reviewApp string) error {
	err := copyNamespace(sourceApp, reviewApp)
	if err != nil {
		return err
	}
	for _, kind := range []string{"roles", "rolebindings"} {
		rolesErr := copyResources(kind, sourceApp, reviewApp)
		if rolesErr != nil {
			return rolesErr
		}
	}
	err = copyResources("secrets", sourceApp, reviewApp, "--field-selector", "type=Opaque")
	if err != nil {
		return err
	}
	return nil
}

// ReviewAppTearDown tears down a reviewapp (namespace, secrets, configmaps, etc.)
func ReviewAppTearDown(reviewApp string) error {
	err := core.DestroyTuberApp(reviewApp)
	if err != nil {
		return err
	}

	return nil
}

func copyNamespace(sourceApp string, reviewApp string) error {
	resource, err := k8s.Get("namespace", sourceApp, sourceApp, "-o", "json")
	if err != nil {
		return err
	}
	resource, err = duplicateResource(resource, sourceApp, reviewApp)
	if err != nil {
		return err
	}
	err = k8s.Apply(resource, reviewApp)
	if err != nil {
		return err
	}
	return nil
}

func copyResources(kind string, sourceApp string, reviewApp string, args ...string) error {
	data, err := duplicatedResources(kind, sourceApp, reviewApp, args...)
	if err != nil {
		return err
	}
	for _, resource := range data {
		applyErr := k8s.Apply(resource, reviewApp)
		if applyErr != nil {
			return applyErr
		}
	}
	return nil
}

func duplicatedResources(kind string, sourceApp string, reviewApp string, args ...string) ([][]byte, error) {
	list, err := k8s.ListKind(kind, sourceApp, args...)
	if err != nil {
		return nil, err
	}
	var resources [][]byte
	for _, resource := range list.Items {
		replicated, replicationErr := duplicateResource(resource, sourceApp, reviewApp)
		if replicationErr != nil {
			return nil, replicationErr
		}
		resources = append(resources, replicated)
	}
	return resources, nil
}

var nonGenericMetadata = []string{"annotations", "creationTimestamp", "namespace", "resourceVersion", "selfLink", "uid"}

func duplicateResource(resource []byte, sourceApp string, reviewApp string) ([]byte, error) {
	unmarshalled := make(map[string]interface{})
	err := json.Unmarshal(resource, &unmarshalled)
	if err != nil {
		return nil, err
	}
	metadata := unmarshalled["metadata"]
	stringKeyMetadata, ok := metadata.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resource metadata could not be coerced into map[string]interface{} for duplication")
	}
	for _, key := range nonGenericMetadata {
		delete(stringKeyMetadata, key)
	}

	stringName, ok := stringKeyMetadata["name"].(string)
	if !ok {
		return nil, fmt.Errorf("resource name could not be coerced into string for potential replacement")
	}
	if strings.Contains(stringName, sourceApp) {
		renamed := strings.ReplaceAll(stringName, sourceApp, reviewApp)
		stringKeyMetadata["name"] = renamed
	}

	unmarshalled["metadata"] = stringKeyMetadata

	genericized, err := json.Marshal(unmarshalled)
	if err != nil {
		return nil, err
	}
	return genericized, nil
}

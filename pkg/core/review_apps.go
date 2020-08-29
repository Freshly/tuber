package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"tuber/pkg/k8s"
)

// NewReviewAppSetup replicates a namespace and its roles, rolebindings, and secrets after removing their non-generic metadata.
// Also renames source app name matches across all relevant resources.
func NewReviewAppSetup(sourceApp string, targetApp string) error {
	err := duplicateNamespace(sourceApp, targetApp)
	if err != nil {
		return err
	}
	for _, kind := range []string{"roles", "rolebindings"} {
		rolesErr := replicateKindToTarget(kind, sourceApp, targetApp)
		if rolesErr != nil {
			return rolesErr
		}
	}
	err = replicateKindToTarget("secrets", sourceApp, targetApp, "--field-selector", "type=Opaque")
	if err != nil {
		return err
	}
	return nil
}

func duplicateNamespace(sourceApp string, targetApp string) error {
	resource, err := k8s.Get("namespace", sourceApp, sourceApp, "-o", "json")
	if err != nil {
		return err
	}
	resource, err = replicateResource(resource, sourceApp, targetApp)
	if err != nil {
		return err
	}
	err = k8s.Apply(resource, targetApp)
	if err != nil {
		return err
	}
	return nil
}

func replicateKindToTarget(kind string, sourceApp string, targetApp string, args ...string) error {
	data, err := replicatedByKind(kind, sourceApp, targetApp, args...)
	if err != nil {
		return err
	}
	for _, resource := range data {
		applyErr := k8s.Apply(resource, targetApp)
		if applyErr != nil {
			return applyErr
		}
	}
	return nil
}

func replicatedByKind(kind string, sourceApp string, targetApp string, args ...string) ([][]byte, error) {
	list, err := k8s.ListKind(kind, sourceApp, args...)
	if err != nil {
		return nil, err
	}
	var resources [][]byte
	for _, resource := range list.Items {
		replicated, replicationErr := replicateResource(resource, sourceApp, targetApp)
		if replicationErr != nil {
			return nil, replicationErr
		}
		resources = append(resources, replicated)
	}
	return resources, nil
}

var nonGenericMetadata = []string{"annotations", "creationTimestamp", "namespace", "resourceVersion", "selfLink", "uid"}

func replicateResource(resource []byte, sourceApp string, targetApp string) ([]byte, error) {
	unmarshalled := make(map[string]interface{})
	err := json.Unmarshal(resource, &unmarshalled)
	if err != nil {
		return nil, err
	}
	metadata := unmarshalled["metadata"]
	stringKeyMetadata, ok := metadata.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resource metadata could not be coerced into map[string]interface{} for genericization")
	}
	for _, key := range nonGenericMetadata {
		delete(stringKeyMetadata, key)
	}

	stringName, ok := stringKeyMetadata["name"].(string)
	if !ok {
		return nil, fmt.Errorf("resource name could not be coerced into string for potential replacement")
	}
	if strings.Contains(stringName, sourceApp) {
		renamed := strings.ReplaceAll(stringName, sourceApp, targetApp)
		stringKeyMetadata["name"] = renamed
	}

	unmarshalled["metadata"] = stringKeyMetadata

	genericized, err := json.Marshal(unmarshalled)
	if err != nil {
		return nil, err
	}
	return genericized, nil
}

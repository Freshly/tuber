package k8s

import (
	"encoding/json"
	"strings"
)

type k8sMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type k8sConfigResource struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   k8sMetadata       `json:"metadata"`
	Type       string            `json:"type,omitempty"`
	Data       map[string]string `json:"data"`
	StringData map[string]string `json:"stringData,omitempty"`
}

// ConfigResource represents the editable portion of a configmap
type ConfigResource struct {
	config *k8sConfigResource
	Data   map[string]string
}

// Save persists updates to a configmap to k8s
func (c *ConfigResource) Save(namespace string) (err error) {
	config := c.config
	config.Data = c.Data

	var jsondata []byte
	jsondata, err = json.Marshal(config)

	if err != nil {
		return
	}

	Apply(jsondata, namespace)

	return
}

// GetConfigResource returns a ConfigResource struct with a Data element containing config map entries
func GetConfigResource(name string, namespace string, kind string) (config *ConfigResource, err error) {
	result, err := Get(strings.ToLower(kind), name, namespace, "-o", "json")

	if err != nil {
		return
	}
	var k8sc k8sConfigResource

	if result == nil {
		k8sc = k8sConfigResource{
			APIVersion: "v1",
			Kind:       kind,
			Data:       map[string]string{},
			Metadata:   k8sMetadata{Name: name, Namespace: namespace},
		}
	} else {
		json.Unmarshal(result, &k8sc)
	}

	data := k8sc.Data
	if data == nil {
		data = map[string]string{}
	}

	config = &ConfigResource{
		config: &k8sc,
		Data:   data,
	}
	return
}

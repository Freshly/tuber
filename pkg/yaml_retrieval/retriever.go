package yaml_retrieval

import "tuber/pkg/util"

type Retriever struct {
	Image ImageInfo
	AuthResponse *registryToken
	Layers []layer
}

// ImageInfo contains identifying information about the target image for a tuber apply
type ImageInfo struct {
	Name string
	Tag string
}

// RetrieveAll returns all Yamls for the env-configured registry
func (r Retriever) RetrieveAll() ([]util.Yaml, error) {
	authResponse, err := r.getToken()
	if err != nil {
		return nil, err
	}
	r.AuthResponse = authResponse

	layers, err := r.getLayers()
	if err != nil {
		return nil, err
	}
	r.Layers = layers

	yamls, err := r.collectYamls()
	if err != nil {
		return nil, err
	}

	return yamls, nil
}

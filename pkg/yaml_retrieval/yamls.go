package yaml_retrieval

import (
	"fmt"
	"log"
	"tuber/pkg/util"
)

const megabyte = 1_000_000
const maxSize = megabyte * 1

// RetrieveAll returns all Yamls for the env-configured registry
func RetrieveAll(image util.ImageInfo) ([]util.Yaml, error) {
	authResponse, err := getToken(image)
	if err != nil {
		return nil, err
	}

	layers, err := getLayers(image, authResponse)
	if err != nil {
		return nil, err
	}

	yamls, err := collectYamls(image, authResponse, layers)
	if err != nil {
		return nil, err
	}

	return yamls, nil
}

func collectYamls(image util.ImageInfo, authResponse *registryToken, layers []layer) ([]util.Yaml, error) {
	for _, layer := range layers {
		if layer.Size > maxSize {
			log.Println("Layer too large, skipping...")
			continue
		}

		yamls, err := downloadLayer(image, authResponse, &layer)

		if err != nil {
			switch err.(type) {
			case *notTuberLayerError:
				continue
			default:
				return nil, err
			}
		}

		return yamls, nil
	}

	return nil, fmt.Errorf("no tuber layer found")
}
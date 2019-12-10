package yaml_retrieval

import (
	"fmt"
	"log"
	"tuber/pkg/util"
)

const megabyte = 1_000_000
const maxSize = megabyte * 1

func (r Retriever) collectYamls() ([]util.Yaml, error) {
	for _, layer := range r.Layers {
		if layer.Size > maxSize {
			log.Println("Layer too large, skipping...")
			continue
		}

		yamls, err := r.downloadLayer(&layer)

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
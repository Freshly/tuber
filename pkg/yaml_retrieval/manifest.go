package yaml_retrieval

import (
	"fmt"
	"tuber/pkg/util"
)

const megabyte = 1_000_000
const maxSize = megabyte * 1

type manifest struct {
	Layers []layer `json:"layers"`
	image authorizedImage
}

func (m manifest) downloadYamls() (yamls []util.Yaml, err error) {
	for _, layer := range m.Layers {
		if layer.Size > maxSize {
			continue
		}

		yamls, err := layer.download(m.image)

		if err != nil {
			switch err.(type) {
			case *notTuberLayerError:
				continue
			default:
				return yamls, err
			}
		}

		return yamls, nil
	}
	err = fmt.Errorf("no tuber layer found")
	return
}
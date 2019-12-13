package release

import (
	"tuber/pkg/apply"
	"tuber/pkg/yamldownloader"
)

func New(name string, tag string, token string) (out []byte, err error) {
	var registry = yamldownloader.NewGoogleRegistry(token)
	repository, err := registry.GetRepository(name, "pull")

	if err != nil {
		return
	}

	yamls, err := repository.FindLayer(tag)

	if err != nil {
		return
	}

	out, err = apply.Apply(yamls)

	if err != nil {
		return
	}

	return
}
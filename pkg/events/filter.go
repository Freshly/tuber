package events

import (
	"tuber/pkg/core"
	"tuber/pkg/listener"
)

func filter(e *listener.RegistryEvent) ([]core.TuberApp, error) {
	apps, err := core.TuberApps()
	var filteredApps []core.TuberApp

	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.ImageTag == e.Tag {
			filteredApps = append(filteredApps, app)
		}
	}

	return filteredApps, nil
}

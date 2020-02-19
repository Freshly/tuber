package core

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"tuber/pkg/containers"

	"tuber/pkg/k8s"
)

const tuberConfig = "tuber-apps"

// TuberApp type for Tuber app
type TuberApp struct {
	Tag      string
	ImageTag string
	RepoPath string
	RepoHost string
	Name     string
}

// GetRepositoryLocation returns a RepositoryLocation struct for a given Tuber App
func (ta *TuberApp) GetRepositoryLocation() containers.RepositoryLocation {
	return containers.RepositoryLocation{
		Host: ta.RepoHost,
		Path: ta.RepoPath,
		Tag:  ta.Tag,
	}
}

type appsCache struct {
	apps   []TuberApp
	expiry time.Time
}

var cache *appsCache
var mutex sync.Mutex

func (a appsCache) isExpired() bool {
	return cache.expiry.Before(time.Now())
}

func refreshAppsCache(apps []TuberApp) {
	expiry := time.Now().Add(time.Minute * 5)
	cache = &appsCache{apps: apps, expiry: expiry}
}

func getTuberApps() (apps []TuberApp, err error) {
	config, err := k8s.GetConfig(tuberConfig, "tuber", "ConfigMap")

	if err != nil {
		return
	}

	for name, imageTag := range config.Data {
		split := strings.SplitN(imageTag, ":", 2)
		repoSplit := strings.SplitN(split[0], "/", 2)

		apps = append(apps, TuberApp{
			Name:     name,
			ImageTag: imageTag,
			Tag:      split[1],
			RepoPath: repoSplit[1],
			RepoHost: repoSplit[0],
		})
	}

	return
}

// AppList is a slice of TuberApp structs
type AppList []TuberApp

// FindApp locates a Tuber app within an app-list
func (ta AppList) FindApp(name string) (foundApp *TuberApp, err error) {
	for _, app := range ta {
		if app.Name == name {
			foundApp = &app
			return
		}
	}

	err = fmt.Errorf("app '%s' not found", name)
	return
}

// TuberApps returns a list of Tuber apps
func TuberApps() (apps AppList, err error) {
	mutex.Lock()
	defer mutex.Unlock()
	if cache == nil || cache.isExpired() {
		apps, err = getTuberApps()

		if err == nil {
			refreshAppsCache(apps)
		}
		return
	}

	apps = cache.apps
	return
}

// AddAppConfig add a new configuration to Tuber's config map
func AddAppConfig(appName string, repo string, tag string) (err error) {
	key := appName
	value := fmt.Sprintf("%s:%s", repo, tag)

	return k8s.PatchConfigMap(tuberConfig, "tuber", key, value)
}

// RemoveAppConfig removes a configuration from Tuber's config map
func RemoveAppConfig(appName string) (err error) {
	return k8s.RemoveConfigMapEntry(tuberConfig, "tuber", appName)
}

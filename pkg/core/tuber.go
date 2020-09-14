package core

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"tuber/pkg/containers"

	"tuber/pkg/k8s"
)

const tuberSourceApps = "tuber-apps"
const tuberReviewApps = "tuber-review-apps"

// TuberApp type for Tuber app
type TuberApp struct {
	Tag      string
	ImageTag string
	Repo     string
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

var sourceAppsCache *appsCache
var sourceAppsMutex sync.Mutex

var reviewAppsCache *appsCache
var reviewMutex sync.Mutex

func (a appsCache) isExpired() bool {
	return sourceAppsCache.expiry.Before(time.Now())
}

func refreshSourceAppsCache(apps []TuberApp) {
	expiry := time.Now().Add(time.Second * 10)
	sourceAppsCache = &appsCache{apps: apps, expiry: expiry}
}

func refreshReviewAppsCache(apps []TuberApp) {
	expiry := time.Now().Add(time.Second * 10)
	reviewAppsCache = &appsCache{apps: apps, expiry: expiry}
}

// getTuberApps retrieves data to be stored in sourceAppsCache.
// always use TuberSourceApps function, never this one directly
func getTuberApps(mapname string) (apps []TuberApp, err error) {
	config, err := k8s.GetConfigResource(mapname, "tuber", "ConfigMap")

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
			Repo:     split[0],
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

func FindApp(name string) (foundApp *TuberApp, err error) {
	apps, err := TuberSourceApps()

	if err != nil {
		return
	}

	return apps.FindApp(name)
}

// TuberSourceApps returns a list of Tuber apps
func TuberSourceApps() (apps AppList, err error) {
	sourceAppsMutex.Lock()
	defer sourceAppsMutex.Unlock()
	if sourceAppsCache == nil || sourceAppsCache.isExpired() {
		apps, err = getTuberApps(tuberSourceApps)

		if err == nil {
			refreshSourceAppsCache(apps)
		}
		return
	}

	apps = sourceAppsCache.apps
	return
}

// TuberReviewApps returns a list of Tuber apps
func TuberReviewApps() (apps AppList, err error) {
	reviewMutex.Lock()
	defer reviewMutex.Unlock()
	if reviewAppsCache == nil || reviewAppsCache.isExpired() {
		apps, err = getTuberApps(tuberReviewApps)

		if err == nil {
			refreshReviewAppsCache(apps)
		}
		return
	}

	apps = reviewAppsCache.apps
	return
}

func SourceAndReviewApps() (AppList, error) {
	apps, err := TuberSourceApps()
	if err != nil {
		return AppList{}, err
	}

	reviewApps, err := TuberReviewApps()
	if err != nil {
		return AppList{}, err
	}

	return append(reviewApps, apps...), nil
}

// AddSourceAppConfig adds a new Source app that tuber will monitor and deploy
func AddSourceAppConfig(appName string, repo string, tag string) error {
	return k8s.PatchConfigMap(tuberSourceApps, "tuber", appName, fmt.Sprintf("%s:%s", repo, tag))
}

// AddReviewAppConfig adds a new Review app that tuber will monitor and deploy
func AddReviewAppConfig(appName string, repo string, tag string) error {
	return k8s.PatchConfigMap(tuberReviewApps, "tuber", appName, fmt.Sprintf("%s:%s", repo, tag))
}

// RemoveSourceAppConfig removes a configuration from Tuber's control
func RemoveSourceAppConfig(appName string) (err error) {
	return k8s.RemoveConfigMapEntry(tuberSourceApps, "tuber", appName)
}

// RemoveReviewAppConfig removes a configuration from Tuber's control
func RemoveReviewAppConfig(appName string) (err error) {
	return k8s.RemoveConfigMapEntry(tuberReviewApps, "tuber", appName)
}

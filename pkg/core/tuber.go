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
	Tag       string
	ImageTag  string
	Repo      string
	RepoPath  string
	RepoHost  string
	Name      string
	ReviewApp bool
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

func toTuberApps(data map[string]string, reviewApps bool) ([]TuberApp, error) {
	var apps []TuberApp
	for name, imageTag := range data {
		split := strings.SplitN(imageTag, ":", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("error parsing tuber app %s", name)
		}

		repoSplit := strings.SplitN(split[0], "/", 2)
		if len(repoSplit) != 2 {
			return nil, fmt.Errorf("error parsing tuber app %s", name)
		}
		apps = append(apps, TuberApp{
			Name:      name,
			ImageTag:  imageTag,
			Tag:       split[1],
			Repo:      split[0],
			RepoPath:  repoSplit[1],
			RepoHost:  repoSplit[0],
			ReviewApp: reviewApps,
		})
	}
	return apps, nil
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

func FindReviewApp(name string) (foundApp *TuberApp, err error) {
	apps, err := TuberReviewApps()

	if err != nil {
		return
	}

	return apps.FindApp(name)
}

// TuberSourceApps returns a list of Tuber apps
func TuberSourceApps() (AppList, error) {
	sourceAppsMutex.Lock()
	defer sourceAppsMutex.Unlock()
	if sourceAppsCache == nil || sourceAppsCache.isExpired() {
		config, err := k8s.GetConfigResource(tuberSourceApps, "tuber", "ConfigMap")
		if err != nil {
			return nil, err
		}

		apps, err := toTuberApps(config.Data, false)

		if err != nil {
			return nil, err
		}

		refreshSourceAppsCache(apps)
		return apps, nil
	}

	return sourceAppsCache.apps, nil
}

// TuberReviewApps returns a list of Tuber apps
func TuberReviewApps() (AppList, error) {
	reviewMutex.Lock()
	defer reviewMutex.Unlock()
	if reviewAppsCache == nil || reviewAppsCache.isExpired() {
		config, err := k8s.GetConfigResource(tuberReviewApps, "tuber", "ConfigMap")
		if err != nil {
			return nil, err
		}
		apps, err := toTuberApps(config.Data, true)
		if err != nil {
			return nil, err
		}

		refreshReviewAppsCache(apps)
		return apps, nil
	}

	return reviewAppsCache.apps, nil
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

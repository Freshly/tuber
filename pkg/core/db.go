package core

import (
	"fmt"
	"strconv"

	bolt "go.etcd.io/bbolt"
)

type Data struct {
	db *bolt.DB
}

func NewData(db *bolt.DB) Data {
	return Data{db}
}

func appsBucket(tx *bolt.Tx) *bolt.Bucket {
	if tx == nil {
		return nil
	}
	return tx.Bucket([]byte("apps"))
}

func getNestedBucket(b *bolt.Bucket, key string) *bolt.Bucket {
	if b == nil {
		return nil
	}
	return b.Bucket([]byte(key))
}

func getString(bucket *bolt.Bucket, key string) string {
	return string(bucket.Get([]byte(key)))
}

func setString(bucket *bolt.Bucket, key string, value string) {
	if bucket == nil {
		return
	}
	bucket.Put([]byte(key), []byte(value))
}

func getBool(bucket *bolt.Bucket, key string) bool {
	v, err := strconv.ParseBool(string(bucket.Get([]byte(key))))
	if err != nil {
		return false
	}
	return v
}

func setBool(bucket *bolt.Bucket, key string, value bool) {
	if bucket == nil {
		return
	}
	bucket.Put([]byte(key), []byte(strconv.FormatBool(value)))
}

func (d *Data) toTuberApp(appBucket *bolt.Bucket, appName string) (*TuberApp, error) {
	if appBucket == nil {
		return nil, fmt.Errorf("bolt app not found")
	}
	app := TuberApp{
		Tag:             getString(appBucket, "tag"),
		ImageTag:        getString(appBucket, "imageTag"),
		Repo:            getString(appBucket, "repo"),
		RepoPath:        getString(appBucket, "repoPath"),
		RepoHost:        getString(appBucket, "repoHost"),
		Name:            appName,
		ReviewApp:       getBool(appBucket, "isReviewApp"),
		Paused:          getBool(appBucket, "isPaused"),
		CloudSourceRepo: getString(appBucket, "cloudSourceRepo"),
		SlackChannel:    getString(appBucket, "slackChannel"),
		TriggerID:       getString(appBucket, "triggerID"),
		SourceAppName:   getString(appBucket, "sourceAppName"),
		d:               d,
	}

	state := State{}
	stateBucket := getNestedBucket(appBucket, "state")
	if stateBucket != nil {
		var current []Resource
		currentStateBucket := getNestedBucket(stateBucket, "current")
		currentStateBucket.ForEach(func(k []byte, v []byte) error {
			current = append(current,
				Resource{
					Encoded: getString(currentStateBucket, "encoded"),
					Kind:    getString(currentStateBucket, "kind"),
					Name:    getString(currentStateBucket, "name"),
				},
			)
			return nil
		})
		state.Current = current

		previousStateBucket := getNestedBucket(stateBucket, "previous")
		previousStateBucket.ForEach(func(k []byte, v []byte) error {
			current = append(current,
				Resource{
					Encoded: getString(previousStateBucket, "encoded"),
					Kind:    getString(previousStateBucket, "kind"),
					Name:    getString(previousStateBucket, "name"),
				},
			)
			return nil
		})
		app.State = &state
	}

	varsBucket := getNestedBucket(appBucket, "vars")
	if varsBucket != nil {
		vars := make(map[string]string)
		_ = varsBucket.ForEach(func(k []byte, value []byte) error {
			vars[string(k)] = string(value)
			return nil
		})
		app.Vars = vars
	}

	reviewAppsConfigBucket := getNestedBucket(appBucket, "reviewAppsConfig")
	if reviewAppsConfigBucket != nil {
		var rac ReviewAppsConfig
		rac.Enabled = getBool(reviewAppsConfigBucket, "enabled")
		racVarsBucket := getNestedBucket(reviewAppsConfigBucket, "vars")
		if racVarsBucket != nil {
			vars := make(map[string]string)
			_ = racVarsBucket.ForEach(func(k []byte, value []byte) error {
				vars[string(k)] = string(value)
				return nil
			})
			rac.Vars = vars
		}
		skips := getNestedBucket(reviewAppsConfigBucket, "skips")
		if skips != nil {
			var skipsResources []Resource
			skipsResources = append(skipsResources,
				Resource{
					Kind: getString(skips, "kind"),
					Name: getString(skips, "name"),
				},
			)
			rac.Skips = skipsResources
		}
		app.ReviewAppsConfig = &rac
	}

	return &app, nil
}

func (d *Data) App(appName string) (*TuberApp, error) {
	var app *TuberApp
	err := d.db.View(func(tx *bolt.Tx) error {
		apps := appsBucket(tx)
		if apps == nil {
			return fmt.Errorf("bolt bucket apps not found, this should never happen")
		}
		appBucket := getNestedBucket(apps, appName)
		var err error
		app, err = d.toTuberApp(appBucket, appName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (d *Data) AppsForTag(tag string) ([]*TuberApp, error) {
	var matchingApps []*TuberApp
	err := d.db.View(func(tx *bolt.Tx) error {

		apps := appsBucket(tx)
		if apps == nil {
			return fmt.Errorf("bolt bucket apps not found, this should never happen")
		}

		err := apps.ForEach(func(k []byte, v []byte) error {
			appBucket := getNestedBucket(apps, string(k))
			if getString(appBucket, "imageTag") == tag {
				app, err := d.toTuberApp(appBucket, string(k))
				if err != nil {
					return err
				}
				matchingApps = append(matchingApps, app)
			}
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return matchingApps, nil
}

func (app *TuberApp) Save() error {
	return app.d.saveApp(app)
}

func (d *Data) CreateApp(app *TuberApp) error {
	return d.saveApp(app)
}

func (d *Data) saveApp(app *TuberApp) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		apps := appsBucket(tx)
		if apps == nil {
			return fmt.Errorf("bolt bucket apps not found, this should never happen")
		}
		appBucket, err := apps.CreateBucketIfNotExists([]byte(app.Name))
		if err != nil {
			return err
		}
		setString(appBucket, "tag", app.Tag)
		setString(appBucket, "imageTag", app.ImageTag)
		setString(appBucket, "repo", app.Repo)
		setString(appBucket, "repoPath", app.RepoPath)
		setString(appBucket, "repoHost", app.RepoHost)
		setBool(appBucket, "isReviewApp", app.ReviewApp)
		setBool(appBucket, "isPaused", app.Paused)
		setString(appBucket, "cloudSourceRepo", app.CloudSourceRepo)
		setString(appBucket, "slackChannel", app.SlackChannel)
		setString(appBucket, "triggerID", app.TriggerID)
		setString(appBucket, "sourceAppName", app.SourceAppName)

		stateResourcesBucket, err := appBucket.CreateBucketIfNotExists([]byte("state"))
		if err != nil {
			return err
		}
		currentStateBucket, err := stateResourcesBucket.CreateBucketIfNotExists([]byte("current"))
		if err != nil {
			return err
		}
		previousStateBucket, err := stateResourcesBucket.CreateBucketIfNotExists([]byte("previous"))
		if err != nil {
			return err
		}
		for _, stateResource := range app.State.Current {
			srb, err := currentStateBucket.CreateBucketIfNotExists([]byte(stateResource.Name + stateResource.Kind))
			if err != nil {
				return err
			}
			setString(srb, "encoded", stateResource.Encoded)
			setString(srb, "kind", stateResource.Kind)
			setString(srb, "name", stateResource.Name)
		}

		for _, stateResource := range app.State.Previous {
			srb, err := previousStateBucket.CreateBucketIfNotExists([]byte(stateResource.Name + stateResource.Kind))
			if err != nil {
				return err
			}
			setString(srb, "encoded", stateResource.Encoded)
			setString(srb, "kind", stateResource.Kind)
			setString(srb, "name", stateResource.Name)
		}

		varsBucket, err := appBucket.CreateBucketIfNotExists([]byte("vars"))
		if err != nil {
			return err
		}
		for k, v := range app.Vars {
			setString(varsBucket, k, v)
		}

		if !app.ReviewApp {
			reviewAppsConfigBucket, err := appBucket.CreateBucketIfNotExists([]byte("reviewAppsConfig"))
			if err != nil {
				return err
			}
			setBool(reviewAppsConfigBucket, "enabled", app.ReviewAppsConfig.Enabled)
			racVars, err := reviewAppsConfigBucket.CreateBucketIfNotExists([]byte("vars"))
			if err != nil {
				return err
			}
			for k, v := range app.ReviewAppsConfig.Vars {
				setString(racVars, k, v)
			}
			skips, err := reviewAppsConfigBucket.CreateBucketIfNotExists([]byte("skips"))
			if err != nil {
				return err
			}
			for _, skip := range app.ReviewAppsConfig.Skips {
				skipBucket, err := skips.CreateBucketIfNotExists([]byte(skip.Name + skip.Kind))
				if err != nil {
					return err
				}
				setString(skipBucket, "name", skip.Name)
				setString(skipBucket, "kind", skip.Kind)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

type TuberApp struct {
	Tag              string
	ImageTag         string
	Repo             string
	RepoPath         string
	RepoHost         string
	Name             string
	ReviewApp        bool
	Paused           bool
	CloudSourceRepo  string
	SlackChannel     string
	TriggerID        string
	State            *State
	Vars             map[string]string
	ReviewAppsConfig *ReviewAppsConfig
	SourceAppName    string
	d                *Data
}

type State struct {
	Current  []Resource
	Previous []Resource
}

type Resource struct {
	Encoded string
	Kind    string
	Name    string
}

type ReviewAppsConfig struct {
	Enabled bool
	Vars    map[string]string
	Skips   []Resource
}

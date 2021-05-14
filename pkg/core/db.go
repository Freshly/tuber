package core

import (
	"fmt"
	"strconv"

	"github.com/freshly/tuber/graph/model"
	bolt "go.etcd.io/bbolt"
)

type Data struct {
	db *bolt.DB
}

func NewData(db *bolt.DB) Data {
	return Data{db}
}

func (d *Data) Close() {
	d.db.Close()
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

type NotFoundError struct {
	error
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("bolt app not found")
}

func (d *Data) toTuberApp(appBucket *bolt.Bucket, appName string) (*model.TuberApp, error) {
	if appBucket == nil {
		return nil, NotFoundError{}
	}
	sourceAppName := getString(appBucket, "sourceAppName")
	app := model.TuberApp{
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
		SourceAppName:   sourceAppName,
	}

	state := model.State{}
	stateBucket := getNestedBucket(appBucket, "state")
	if stateBucket != nil {
		var current []*model.Resource
		currentStateBucket := getNestedBucket(stateBucket, "current")
		currentStateBucket.ForEach(func(k []byte, v []byte) error {
			current = append(current,
				&model.Resource{
					Encoded: getString(currentStateBucket, "encoded"),
					Kind:    getString(currentStateBucket, "kind"),
					Name:    getString(currentStateBucket, "name"),
				},
			)
			return nil
		})
		state.Current = current

		var previous []*model.Resource
		previousStateBucket := getNestedBucket(stateBucket, "previous")
		previousStateBucket.ForEach(func(k []byte, v []byte) error {
			previous = append(current,
				&model.Resource{
					Encoded: getString(previousStateBucket, "encoded"),
					Kind:    getString(previousStateBucket, "kind"),
					Name:    getString(previousStateBucket, "name"),
				},
			)
			return nil
		})
		state.Previous = previous
		app.State = &state
	}

	varsBucket := getNestedBucket(appBucket, "vars")
	var vars []*model.Tuple
	if varsBucket != nil {
		_ = varsBucket.ForEach(func(k []byte, value []byte) error {
			vars = append(vars, &model.Tuple{
				Key:   string(k),
				Value: string(value),
			})
			return nil
		})
		app.Vars = vars
	}

	reviewAppsConfigBucket := getNestedBucket(appBucket, "reviewAppsConfig")
	if reviewAppsConfigBucket != nil {
		var rac model.ReviewAppsConfig
		rac.Enabled = getBool(reviewAppsConfigBucket, "enabled")
		racVarsBucket := getNestedBucket(reviewAppsConfigBucket, "vars")
		if racVarsBucket != nil {
			var vars []*model.Tuple
			_ = racVarsBucket.ForEach(func(k []byte, value []byte) error {
				vars = append(vars, &model.Tuple{
					Key:   string(k),
					Value: string(value),
				})
				return nil
			})
			rac.Vars = vars
		}
		skips := getNestedBucket(reviewAppsConfigBucket, "skips")
		if skips != nil {
			var skipsResources []*model.Resource
			skipsResources = append(skipsResources,
				&model.Resource{
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

func (d *Data) ReviewAppsFor(app *model.TuberApp) ([]*model.TuberApp, error) {
	var reviewApps []*model.TuberApp
	err := d.db.View(func(tx *bolt.Tx) error {
		apps := appsBucket(tx)
		err := apps.ForEach(func(k []byte, v []byte) error {
			appBucket := getNestedBucket(apps, string(k))
			if getString(appBucket, "sourceAppName") == app.Name {
				app, err := d.toTuberApp(appBucket, string(k))
				if err != nil {
					return err
				}
				reviewApps = append(reviewApps, app)
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
	return reviewApps, nil
}

func (d *Data) SourceAppFor(app *model.TuberApp) (*model.TuberApp, error) {
	var sourceApp *model.TuberApp
	err := d.db.View(func(tx *bolt.Tx) error {
		apps := appsBucket(tx)
		appBucket := getNestedBucket(apps, app.SourceAppName)
		var err error
		app, err = d.toTuberApp(appBucket, app.SourceAppName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sourceApp, nil
}

// this is the most duplicative thing ive ever done but whatever the signature is fixed
func (d *Data) SourceApps(tag string) ([]*model.TuberApp, error) {
	var matchingApps []*model.TuberApp
	err := d.db.View(func(tx *bolt.Tx) error {

		apps := appsBucket(tx)

		err := apps.ForEach(func(k []byte, v []byte) error {
			appBucket := getNestedBucket(apps, string(k))
			if getString(appBucket, "sourceAppName") == "" {
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

// this is the most duplicative thing ive ever done but whatever the signature is fixed
func (d *Data) ReviewApps(tag string) ([]*model.TuberApp, error) {
	var matchingApps []*model.TuberApp
	err := d.db.View(func(tx *bolt.Tx) error {

		apps := appsBucket(tx)

		err := apps.ForEach(func(k []byte, v []byte) error {
			appBucket := getNestedBucket(apps, string(k))
			if getString(appBucket, "sourceAppName") != "" {
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

// this is the most duplicative thing ive ever done but whatever the signature is fixed
func (d *Data) Apps() ([]*model.TuberApp, error) {
	var apps []*model.TuberApp
	err := d.db.View(func(tx *bolt.Tx) error {

		all := appsBucket(tx)

		err := all.ForEach(func(k []byte, v []byte) error {
			appBucket := getNestedBucket(all, string(k))
			app, err := d.toTuberApp(appBucket, string(k))
			if err != nil {
				return err
			}
			apps = append(apps, app)
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
	return apps, nil
}

func (d *Data) AppExists(appName string) (bool, error) {
	var exists bool
	err := d.db.View(func(tx *bolt.Tx) error {
		apps := appsBucket(tx)
		appBucket := getNestedBucket(apps, appName)
		if appBucket != nil {
			exists = true
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (d *Data) App(appName string) (*model.TuberApp, error) {
	var app *model.TuberApp
	err := d.db.View(func(tx *bolt.Tx) error {
		apps := appsBucket(tx)
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

func (d *Data) AppsForTag(tag string) ([]*model.TuberApp, error) {
	var matchingApps []*model.TuberApp
	err := d.db.View(func(tx *bolt.Tx) error {

		apps := appsBucket(tx)

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

func (d *Data) Save(app *model.TuberApp) error {
	return d.saveApp(app)
}

func (d *Data) DeleteTuberApp(app *model.TuberApp) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		apps := appsBucket(tx)
		err := apps.DeleteBucket([]byte(app.Name))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *Data) saveApp(app *model.TuberApp) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		apps := appsBucket(tx)
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
		for _, tuple := range app.Vars {
			setString(varsBucket, tuple.Key, tuple.Value)
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
			for _, tuple := range app.ReviewAppsConfig.Vars {
				setString(racVars, tuple.Key, tuple.Value)
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

package core

import (
	"fmt"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/db"
)

type DB struct {
	db *db.DB
}

func NewDB(db *db.DB) *DB {
	return &DB{db: db}
}

func (d *DB) Close() {
	d.db.Close()
}

func (d *DB) ReviewAppsFor(app *model.TuberApp) ([]*model.TuberApp, error) {
	r, err := d.db.Get(model.TuberApp{}, db.Q().String("sourceAppName", app.Name).Bool("reviewApp", true))
	if err != nil {
		return nil, err
	}
	return assertAll(r)
}

func (d *DB) SourceAppFor(app *model.TuberApp) (*model.TuberApp, error) {
	r, err := d.db.Find(model.TuberApp{}, app.SourceAppName)
	if err != nil {
		return nil, err
	}
	return assert(r)
}

func (d *DB) SourceApps() ([]*model.TuberApp, error) {
	r, err := d.db.Get(model.TuberApp{}, db.Q().Bool("reviewApp", false))
	if err != nil {
		return nil, err
	}
	return assertAll(r)
}

func (d *DB) ReviewApps() ([]*model.TuberApp, error) {
	r, err := d.db.Get(model.TuberApp{}, db.Q().Bool("reviewApp", true))
	if err != nil {
		return nil, err
	}
	return assertAll(r)
}

func (d *DB) Apps() ([]*model.TuberApp, error) {
	r, err := d.db.Get(model.TuberApp{}, db.Query{})
	if err != nil {
		return nil, err
	}
	return assertAll(r)
}

func (d *DB) AppExists(appName string) bool {
	return d.db.Exists(model.TuberApp{}, appName)
}

func (d *DB) AppsForTag(tag string) ([]*model.TuberApp, error) {
	r, err := d.db.Get(model.TuberApp{}, db.Q().String("imageTag", tag))
	if err != nil {
		return nil, err
	}
	return assertAll(r)
}

func (d *DB) App(appName string) (*model.TuberApp, error) {
	r, err := d.db.Find(model.TuberApp{}, appName)
	if err != nil {
		return nil, err
	}

	return assert(r)
}

func (d *DB) SaveApp(app *model.TuberApp) error {
	return d.db.Save(app)
}

func (d *DB) DeleteApp(app *model.TuberApp) error {
	return d.db.Delete(app, app.Name)
}

func assert(m db.Model) (*model.TuberApp, error) {
	app, ok := m.(model.TuberApp)
	if !ok {
		return nil, fmt.Errorf("db result could not be asserted as model.TuberApp")
	}
	return &app, nil
}

func assertAll(ms []db.Model) ([]*model.TuberApp, error) {
	var apps []*model.TuberApp
	for _, m := range ms {
		app, ok := m.(model.TuberApp)
		if !ok {
			return nil, fmt.Errorf("db result could not be asserted as model.TuberApp")
		}
		apps = append(apps, &app)
	}
	return apps, nil
}

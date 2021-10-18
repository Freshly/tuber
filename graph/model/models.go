package model

import (
	"encoding/json"
	"time"

	"github.com/freshly/tuber/pkg/db"
)

func (t TuberApp) DBIndexes() (map[string]string, map[string]bool, map[string]int) {
	return map[string]string{
			"name":            t.Name,
			"imageTag":        t.ImageTag,
			"sourceAppName":   t.SourceAppName,
			"cloudSourceRepo": t.CloudSourceRepo,
		}, map[string]bool{
			"reviewApp": t.ReviewApp,
		}, map[string]int{}
}

func (t TuberApp) DBRoot() string {
	return "apps"
}

func (t TuberApp) DBKey() string {
	return t.Name
}

func (t TuberApp) DBMarshal() ([]byte, error) {
	return json.Marshal(t)
}

func (t TuberApp) DBUnmarshal(data []byte) (db.Model, error) {
	var app TuberApp
	err := json.Unmarshal(data, &app)
	if err != nil {
		return nil, err
	}
	if app.State == nil {
		app.State = &State{}
	}
	if app.ReviewAppsConfig == nil {
		app.ReviewAppsConfig = &ReviewAppsConfig{}
	}
	return app, nil
}

func (t TuberApp) TimestampFormat() string {
	return time.RFC1123
}

func (t TuberApp) ParsedCreatedAt() (time.Time, error) {
	parsed, err := time.Parse(t.TimestampFormat(), t.CreatedAt)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}

func (t TuberApp) ParsedUpdatedAt() (time.Time, error) {
	parsed, err := time.Parse(t.TimestampFormat(), t.UpdatedAt)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}

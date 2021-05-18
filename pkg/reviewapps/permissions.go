package reviewapps

import (
	"fmt"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/db"
	"github.com/freshly/tuber/pkg/k8s"

	"go.uber.org/zap"
)

func canCreate(logger *zap.Logger, db *db.DB, appName string, token string) (bool, error) {
	if appName == "tuber" || token == "" {
		return false, nil
	}

	if !appExists(db, appName) {
		return false, nil
	}

	return k8s.CanDeploy(appName, fmt.Sprintf("--token=%s", token))
}

func appExists(db *db.DB, appName string) bool {
	return db.Exists(model.TuberApp{Name: appName})
}

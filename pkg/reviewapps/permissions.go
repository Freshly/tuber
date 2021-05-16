package reviewapps

import (
	"fmt"

	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/k8s"

	"go.uber.org/zap"
)

func canCreate(logger *zap.Logger, db *core.DB, appName string, token string) (bool, error) {
	if appName == "tuber" || token == "" {
		return false, nil
	}

	if !appExists(db, appName) {
		return false, nil
	}

	return k8s.CanDeploy(appName, fmt.Sprintf("--token=%s", token))
}

func appExists(db *core.DB, appName string) bool {
	return db.AppExists(appName)
}

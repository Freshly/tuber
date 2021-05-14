package reviewapps

import (
	"fmt"

	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/k8s"

	"go.uber.org/zap"
)

func canCreate(logger *zap.Logger, db *core.Data, appName string, token string) (bool, error) {
	if appName == "tuber" || token == "" {
		return false, nil
	}

	exists, err := appExists(db, appName)
	if err != nil || !exists {
		return false, err
	}

	return k8s.CanDeploy(appName, fmt.Sprintf("--token=%s", token))
}

func appExists(db *core.Data, appName string) (bool, error) {
	exists, err := db.AppExists(appName)
	if err != nil {
		return false, err
	}
	return exists, nil
}

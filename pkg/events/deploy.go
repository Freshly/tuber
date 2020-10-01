package events

import (
	"fmt"
	"tuber/pkg/containers"
	"tuber/pkg/core"

	"go.uber.org/zap"
)

func deploy(logger zap.Logger, app *core.TuberApp, digest string, creds []byte, clusterData *core.ClusterData) (err error) {
	prereleaseYamls, releaseYamls, err := containers.GetTuberLayer(app.GetRepositoryLocation(), creds)

	if err != nil {
		return
	}

	if len(prereleaseYamls) > 0 {
		logger.Info("prerelease: starting",
			zap.String("event", "begin"),
			zap.String("digest", digest),
		)

		err = core.RunPrerelease(prereleaseYamls, app, digest, clusterData)

		if err != nil {
			err = fmt.Errorf("prerelease error: %s", err.Error())
			return
		}

		logger.Info("prerelease: done",
			zap.String("event", "complete"),
			zap.String("digest", digest),
		)
	}

	releaseIDs, err := core.ReleaseTubers(releaseYamls, app, digest, clusterData)
	if err != nil {
		return
	}
	logger.Debug("Applied tubers with release ids:",
		zap.Strings("releaseIds", releaseIDs),
	)

	return nil
}

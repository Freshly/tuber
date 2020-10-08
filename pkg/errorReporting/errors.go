package errorReporting

import (
	"go.uber.org/zap"
)

type ErrorIntegrations struct {
	Reporters []ErrorReporter
}

type ErrorReporter interface {
	init() error
	reportErr(err error)
	enabled() bool
}

func StartWatching(integrations ErrorIntegrations, logger *zap.Logger) (chan error, error) {
	errorsChannel := make(chan error)
	for _, integration := range integrations.Reporters {
		if integration.enabled() {
			err := integration.init()
			if err != nil {
				return nil, err
			}
		}
	}

	go func() {
		for err := range errorsChannel {
			logger.Warn("nonspecific error", zap.Error(err))
			for _, integration := range integrations.Reporters {
				if integration.enabled() {
					integration.reportErr(err)
				}
			}
		}
	}()

	return errorsChannel, nil
}

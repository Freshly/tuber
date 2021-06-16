package monitor

import (
	"time"

	"go.uber.org/zap"
)

func Sentry(logger *zap.Logger, url string, bearer string, duration time.Duration) (bool, string) {
	timeout := time.Now().Add(duration)
	for {
		logger.Debug("pinging sentry at: " + url)
		if time.Now().After(timeout) {
			return true, ""
		}
		healthy, message := checkSentry(logger, url, bearer)
		if !healthy {
			return false, message
		}
		time.Sleep(30 * time.Second)
	}
}

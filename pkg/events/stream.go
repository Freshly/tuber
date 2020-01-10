package events

import (
	"sync"
	"tuber/pkg/core"
	"tuber/pkg/util"

	"go.uber.org/zap"
)

type streamer struct {
	token       string
	logger      *zap.Logger
	clusterData *core.ClusterData
}

// NewStreamer creates a new Streamer struct
func NewStreamer(token string, logger *zap.Logger, clusterData *core.ClusterData) *streamer {
	return &streamer{token: token, logger: logger, clusterData: clusterData}
}

type failedRelease struct {
	err   error
	event *util.RegistryEvent
}

// Err returns the error causing a failed release
func (f *failedRelease) Err() error { return f.err }

// Event returns the failed release event
func (f *failedRelease) Event() *util.RegistryEvent { return f.event }

// Stream streams a stream
func (s *streamer) Stream(unprocessed <-chan *util.RegistryEvent, processed chan<- *util.RegistryEvent, chErr chan<- util.FailedRelease, chErrReports chan<- error) {
	defer close(processed)
	defer close(chErr)

	var wait = &sync.WaitGroup{}

	for event := range unprocessed {
		go func(event *util.RegistryEvent) {
			var err error

			wait.Add(1)
			defer wait.Done()

			defer func() {
				if err != nil {
					chErr <- &failedRelease{err: err, event: event}
					chErrReports <- err
				} else {
					processed <- event
				}
			}()

			pendingRelease, err := filter(event)

			if err != nil || pendingRelease == nil {
				return
			}

			var releaseLog = s.logger.With(
				zap.String("releaseName", pendingRelease.Name),
				zap.String("releaseBranch", pendingRelease.Tag),
			)

			releaseLog.Info("release: starting")

			output, err := publish(pendingRelease, event.Digest, s.token, s.clusterData)

			if err != nil {
				releaseLog.Warn(
					"release: error",
					zap.Error(err),
					zap.String("output", string(output)),
				)
			} else {
				releaseLog.Info("release: done")
			}
		}(event)
	}

	// Wait for all publish goroutines to be done.
	wait.Wait()
}

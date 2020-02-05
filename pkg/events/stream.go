package events

import (
	"sync"
	"tuber/pkg/core"
	"tuber/pkg/listener"

	"go.uber.org/zap"
)

type streamer struct {
	creds       []byte
	logger      *zap.Logger
	clusterData *core.ClusterData
}

// NewStreamer creates a new Streamer struct
func NewStreamer(creds []byte, logger *zap.Logger, clusterData *core.ClusterData) *streamer {
	return &streamer{creds: creds, logger: logger, clusterData: clusterData}
}

// Stream streams a stream
func (s *streamer) Stream(unprocessed <-chan *listener.RegistryEvent, processed chan<- *listener.RegistryEvent, chErr chan<- listener.FailedRelease, chErrReports chan<- error) {
	defer close(processed)
	defer close(chErr)

	var wait = &sync.WaitGroup{}

	for event := range unprocessed {
		go func(event *listener.RegistryEvent) {
			var err error

			wait.Add(1)
			defer wait.Done()

			defer func() {
				if err != nil {
					chErr <- listener.FailedRelease{Err: err, Event: event}
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

			err = publish(pendingRelease, event.Digest, s.creds, s.clusterData)

			if err != nil {
				releaseLog.Warn(
					"release: error",
					zap.Error(err),
				)
			} else {
				releaseLog.Info("release: done")
			}
		}(event)
	}

	// Wait for all publish goroutines to be done.
	wait.Wait()
}

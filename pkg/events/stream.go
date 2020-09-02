package events

import (
	"strings"
	"sync"
	"time"
	"tuber/pkg/core"
	"tuber/pkg/listener"

	"go.uber.org/zap"
)

type Streamer struct {
	Creds             []byte
	Logger            *zap.Logger
	ClusterData       *core.ClusterData
	ReviewAppsEnabled bool
	Unprocessed       <-chan *listener.RegistryEvent
	Processed         chan<- *listener.RegistryEvent
	ChErr             chan<- listener.FailedRelease
	ChErrReports      chan<- error
}

// Stream streams a stream
func (s Streamer) Stream() {
	defer close(s.Processed)
	defer close(s.ChErr)
	defer close(s.ChErrReports)

	var wait = &sync.WaitGroup{}

	for event := range s.Unprocessed {
		go func(event *listener.RegistryEvent) {
			wait.Add(1)
			defer wait.Done()

			apps, err := s.apps()
			if err != nil {
				s.reportFailedRelease(event, s.Logger, err)
				return
			}

			s.processEvent(event, apps)
		}(event)
	}
	wait.Wait()
}

func (s Streamer) apps() ([]core.TuberApp, error) {
	if s.ReviewAppsEnabled {
		return core.TuberApps()
	} else {
		return core.SourceAndReviewApps()
	}
}

func (s Streamer) processEvent(event *listener.RegistryEvent, apps []core.TuberApp) {
	for _, app := range apps {
		if app.ImageTag == event.Tag {
			s.runDeploy(app, event)
		}
	}
}

func (s Streamer) releaseLogger(app core.TuberApp) *zap.Logger {
	imageTag := strings.Split(app.ImageTag, ":")[1]
	return s.Logger.With(
		zap.String("name", app.Name),
		zap.String("branch", app.Tag),
		zap.String("imageTag", imageTag),
		zap.String("action", "release"),
	)
}

func (s Streamer) runDeploy(app core.TuberApp, event *listener.RegistryEvent) {
	releaseLog := s.releaseLogger(app)

	startTime := time.Now()
	releaseLog.Info("release: starting", zap.String("event", "begin"))

	err := deploy(*releaseLog, &app, event.Digest, s.Creds, s.ClusterData)

	if err != nil {
		s.reportFailedRelease(event, releaseLog, err)
	} else {
		s.reportSuccessfulRelease(event, releaseLog, startTime)
	}
}

func (s Streamer) reportSuccessfulRelease(event *listener.RegistryEvent, releaseLog *zap.Logger, startTime time.Time) {
	releaseLog.Info("release: done", zap.String("event", "complete"), zap.Duration("duration", time.Since(startTime)))
	s.Processed <- event
}

func (s Streamer) reportFailedRelease(event *listener.RegistryEvent, releaseLog *zap.Logger, err error) {
	releaseLog.Warn(
		"release: error",
		zap.String("event", "error"),
		zap.Error(err),
	)
	s.ChErr <- listener.FailedRelease{Err: err, Event: event}
	s.ChErrReports <- err
	s.Processed <- event
}

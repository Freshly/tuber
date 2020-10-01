package events

import (
	"strings"
	"sync"
	"time"
	"tuber/pkg/core"
	"tuber/pkg/listener"

	"go.uber.org/zap"
)

// EventProcessor processes events
type EventProcessor struct {
	Creds             []byte
	Logger            *zap.Logger
	ClusterData       *core.ClusterData
	ReviewAppsEnabled bool
	Unprocessed       <-chan *listener.RegistryEvent
	Processed         chan<- *listener.RegistryEvent
	ChErr             chan<- listener.FailedRelease
	ChErrReports      chan<- error
}

// Start streams a stream
func (p EventProcessor) Start() {
	defer close(p.Processed)
	defer close(p.ChErr)
	defer close(p.ChErrReports)

	p.Logger.Info("Event processor starting", zap.Bool("review apps enabled", p.ReviewAppsEnabled))
	var wait = &sync.WaitGroup{}

	for event := range p.Unprocessed {
		go func(event *listener.RegistryEvent) {
			wait.Add(1)
			defer wait.Done()

			apps, err := p.apps()
			if err != nil {
				p.Logger.Error("An error occured looking up tuber apps! Reporting the event as failed and acking",
					zap.Error(err),
					zap.String("tag", event.Tag),
					zap.String("digest", event.Digest),
				)
				p.reportFailedRelease(event, p.Logger, err)
				return
			}

			p.processEvent(event, apps)
		}(event)
	}
	wait.Wait()
}

func (p EventProcessor) apps() ([]core.TuberApp, error) {
	if p.ReviewAppsEnabled {
		p.Logger.Debug("Listing source and review apps")
		return core.SourceAndReviewApps()
	}

	p.Logger.Debug("Listing source apps")
	return core.TuberSourceApps()
}

func (p EventProcessor) processEvent(event *listener.RegistryEvent, apps []core.TuberApp) {
	p.Logger.Info("Processing event",
		zap.String("tag", event.Tag),
		zap.String("digest", event.Digest),
	)
	p.Logger.Debug("Listing tuber apps",
		zap.Any("apps", apps),
	)

	for _, app := range apps {
		if app.ImageTag == event.Tag {
			p.runDeploy(app, event)
		}
	}
}

func (p EventProcessor) releaseLogger(app core.TuberApp) *zap.Logger {
	imageTag := strings.Split(app.ImageTag, ":")[1]
	return p.Logger.With(
		zap.String("name", app.Name),
		zap.String("branch", app.Tag),
		zap.String("imageTag", imageTag),
		zap.String("action", "release"),
	)
}

func (p EventProcessor) runDeploy(app core.TuberApp, event *listener.RegistryEvent) {
	releaseLog := p.releaseLogger(app)

	startTime := time.Now()
	releaseLog.Info("release: starting",
		zap.String("event", "begin"),
		zap.String("tag", event.Tag),
		zap.String("digest", event.Digest),
	)

	err := deploy(*releaseLog, &app, event.Digest, p.Creds, p.ClusterData)

	if err != nil {
		p.reportFailedRelease(event, releaseLog, err)
	} else {
		p.reportSuccessfulRelease(event, releaseLog, startTime)
	}
}

func (p EventProcessor) reportSuccessfulRelease(event *listener.RegistryEvent, releaseLog *zap.Logger, startTime time.Time) {
	releaseLog.Info("release: done",
		zap.String("event", "complete"),
		zap.Duration("duration", time.Since(startTime)),
		zap.String("tag", event.Tag),
		zap.String("digest", event.Digest),
	)
	p.Processed <- event
}

func (p EventProcessor) reportFailedRelease(event *listener.RegistryEvent, releaseLog *zap.Logger, err error) {
	releaseLog.Warn(
		"release: error",
		zap.String("event", "error"),
		zap.Error(err),
		zap.String("tag", event.Tag),
		zap.String("digest", event.Digest),
	)
	p.ChErr <- listener.FailedRelease{Err: err, Event: event}
	p.ChErrReports <- err
	p.Processed <- event
}

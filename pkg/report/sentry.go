package report

import (
	"github.com/getsentry/sentry-go"
)

type Sentry struct {
	Enable  bool
	Options sentry.ClientOptions
}

func (s Sentry) init() error {
	err := sentry.Init(s.Options)
	if err != nil {
		return err
	}
	defer sentry.Recover()
	return nil
}

func (s Sentry) enabled() bool {
	return s.Enable
}

func (s Sentry) reportErr(err error, scopeData Scope) {
	sentry.WithScope(func(scope *sentry.Scope) {
		for k, v := range scopeData {
			scope.SetTag(k, v)
		}
		sentry.CaptureException(err)
	})
}

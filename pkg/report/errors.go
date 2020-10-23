package report

var ErrorReporters []ErrorReporter

type ErrorReporter interface {
	init() error
	reportErr(err error, scopeData Scope)
	enabled() bool
}

type Scope map[string]string

func InitErrorReporters() error {
	for _, integration := range ErrorReporters {
		if integration.enabled() {
			err := integration.init()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Error(err error, scopeData Scope) {
	for _, integration := range ErrorReporters {
		if integration.enabled() {
			integration.reportErr(err, scopeData)
		}
	}
}

func (s Scope) AddScope(additional Scope) Scope {
	var new = s
	for k, v := range additional {
		new[k] = v
	}
	return new
}

func (s Scope) WithContext(value string) Scope {
	return s.AddScope(Scope{"context": value})
}

package report

// ErrorReporters global set during startup (so env vars can be set) of all reporters
var ErrorReporters []ErrorReporter

// ErrorReporter interface for integrations or custom plugins for reporting errors
type ErrorReporter interface {
	init() error
	reportErr(err error, scopeData Scope)
	enabled() bool
}

// Scope map[string]string of keys and values to be stamped onto reported errors for context
type Scope map[string]string

// InitErrorReporters loops through the enabled reporters and calls init()
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

// Error reports an error to all enabled reporters
func Error(err error, scopeData Scope) {
	for _, integration := range ErrorReporters {
		if integration.enabled() {
			integration.reportErr(err, scopeData)
		}
	}
}

// AddScope combines (overwriting existing keys) two scopes
func (s Scope) AddScope(additional Scope) Scope {
	var new = s
	for k, v := range additional {
		new[k] = v
	}
	return new
}

// WithContext sets a "context" key on an existing scope to a given value
func (s Scope) WithContext(value string) Scope {
	return s.AddScope(Scope{"context": value})
}

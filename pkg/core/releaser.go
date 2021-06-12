package core

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/gcr"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/report"

	"github.com/goccy/go-yaml"
	"go.uber.org/zap"
)

type releaser struct {
	logger           *zap.Logger
	errorScope       report.Scope
	app              *model.TuberApp
	digest           string
	data             *ClusterData
	releaseYamls     []string
	prereleaseYamls  []string
	postreleaseYamls []string
	db               *DB
}

type ErrorContext struct {
	scope   report.Scope
	logger  *zap.Logger
	err     error
	context string
}

func (e ErrorContext) Error() string {
	return e.err.Error()
}

func (r releaser) releaseError(err error) error {
	var context string
	var scope = r.errorScope
	var logger = r.logger

	errorContext, ok := err.(ErrorContext)
	if ok {
		context = errorContext.context
		if errorContext.scope != nil {
			scope = r.errorScope.AddScope(errorContext.scope).WithContext(context)
		}
		if errorContext.logger != nil {
			logger = errorContext.logger
		}
		err = errorContext.err
	} else {
		context = "unknown"
	}

	logger.Error("release error", zap.Error(err), zap.String("context", context))
	report.Error(err, scope)

	return err
}

// Release interpolates and applies an app's resources. It removes deleted resources, and rolls back on any release failure.
// If you edit a resource manually, and a release fails, tuber will roll back to the previously released state of the object, not to the state you manually specified.
func Release(db *DB, yamls *gcr.AppYamls, logger *zap.Logger, errorScope report.Scope, app *model.TuberApp, digest string, data *ClusterData) error {
	return releaser{
		logger:           logger,
		errorScope:       errorScope,
		releaseYamls:     yamls.Release,
		prereleaseYamls:  yamls.Prerelease,
		postreleaseYamls: yamls.PostRelease,
		app:              app,
		digest:           digest,
		data:             data,
		db:               db,
	}.release()
}

func (r releaser) release() error {
	r.logger.Debug("releaser starting")

	prereleaseResources, configs, workloads, postreleaseResources, err := r.resourcesToApply()
	if err != nil {
		return r.releaseError(err)
	}

	decodedStateBeforeApply, err := r.currentState()
	if err != nil {
		return r.releaseError(err)
	}

	if len(prereleaseResources) > 0 {
		r.logger.Debug("prerelease starting")

		err = RunPrerelease(prereleaseResources, r.app)
		if err != nil {
			return ErrorContext{context: "prerelease", err: err}
		}

		r.logger.Debug("prerelease complete")
	}

	appliedConfigs, err := r.apply(configs)
	if err != nil {
		_ = r.releaseError(err)
		r.rollback(appliedConfigs, decodedStateBeforeApply)
		return err
	}

	appliedWorkloads, err := r.apply(workloads)
	if err != nil {
		_ = r.releaseError(err)
		_, configRollbackErrors := r.rollback(appliedConfigs, decodedStateBeforeApply)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, decodedStateBeforeApply)
		for _, rollbackError := range append(configRollbackErrors, workloadRollbackErrors...) {
			_ = r.releaseError(rollbackError)
		}
		watchErrors := r.watchRollback(rolledBackResources)
		for _, watchError := range watchErrors {
			_ = r.releaseError(watchError)
		}
		return err
	}

	err = r.watchWorkloads(appliedWorkloads)
	if err != nil {
		_ = r.releaseError(err)
		_, configRollbackErrors := r.rollback(appliedConfigs, decodedStateBeforeApply)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, decodedStateBeforeApply)
		for _, rollbackError := range append(configRollbackErrors, workloadRollbackErrors...) {
			_ = r.releaseError(rollbackError)
		}
		watchErrors := r.watchRollback(rolledBackResources)
		for _, watchError := range watchErrors {
			_ = r.releaseError(watchError)
		}
		return err
	}

	appliedPostreleaseResources, err := r.apply(postreleaseResources)
	if err != nil {
		_ = r.releaseError(err)
		_, configRollbackErrors := r.rollback(appliedConfigs, decodedStateBeforeApply)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, decodedStateBeforeApply)
		for _, rollbackError := range append(configRollbackErrors, workloadRollbackErrors...) {
			_ = r.releaseError(rollbackError)
		}
		watchErrors := r.watchRollback(rolledBackResources)
		for _, watchError := range watchErrors {
			_ = r.releaseError(watchError)
		}
		return err
	}

	err = r.watchWorkloads(appliedPostreleaseResources)
	if err != nil {
		_ = r.releaseError(err)
		_, configRollbackErrors := r.rollback(appliedConfigs, decodedStateBeforeApply)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, decodedStateBeforeApply)
		rolledBackPostreleaseResources, postreleaseRollbackErrors := r.rollback(appliedPostreleaseResources, decodedStateBeforeApply)
		for _, rollbackError := range append(configRollbackErrors, append(workloadRollbackErrors, postreleaseRollbackErrors...)...) {
			_ = r.releaseError(rollbackError)
		}
		watchErrors := r.watchRollback(append(rolledBackResources, rolledBackPostreleaseResources...))
		for _, watchError := range watchErrors {
			_ = r.releaseError(watchError)
		}
		return err
	}

	err = r.reconcileState(decodedStateBeforeApply, appliedWorkloads, appliedConfigs, appliedPostreleaseResources)
	if err != nil {
		return r.releaseError(err)
	}

	return nil
}

type appResource struct {
	contents        []byte
	kind            string
	name            string
	timeout         time.Duration
	rollbackTimeout time.Duration
	watchUrl        string
	watchDuration   time.Duration
}

func (a appResource) isWorkload() bool {
	return a.supportsRollback() || a.kind == "Pod"
}

func (a appResource) supportsRollback() bool {
	return a.kind == "Deployment" || a.kind == "Daemonset" || a.kind == "StatefulSet"
}

func (a appResource) canBeManaged() bool {
	return a.kind != "Secret" && a.kind != "ClusterRole" && a.kind != "ClusterRoleBinding"
}

func (a appResource) scopes(r releaser) (report.Scope, *zap.Logger) {
	scope := r.errorScope.AddScope(report.Scope{"resourceName": a.name, "resourceKind": a.kind})
	logger := r.logger.With(zap.String("resourceName", a.name), zap.String("resourceKind", a.kind))
	return scope, logger
}

type appResources []appResource

func (a appResources) encode() []*model.Resource {
	var encoded []*model.Resource
	for _, resource := range a {
		m := &model.Resource{
			Kind:    resource.kind,
			Name:    resource.name,
			Encoded: base64.StdEncoding.EncodeToString(resource.contents),
		}
		encoded = append(encoded, m)
	}
	return encoded
}

func (r releaser) currentState() (appResources, error) {
	var resources appResources
	if r.app.State != nil {
		for _, managed := range r.app.State.Current {
			contents, err := base64.StdEncoding.DecodeString(managed.Encoded)
			if err != nil {
				return nil, err
			}
			resources = append(resources, appResource{contents: contents, kind: managed.Kind, name: managed.Name})
		}
	}

	return resources, nil
}

type metadata struct {
	Name        string                 `yaml:"name"`
	Annotations map[string]interface{} `yaml:"annotations"`
}

type parsedResource struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   metadata `yaml:"metadata"`
}

func (r releaser) resourcesToApply() ([]appResource, []appResource, []appResource, []appResource, error) {
	d := releaseData(r.digest, r.app, r.data)

	prereleaseResources, err := r.yamlToAppResource(r.prereleaseYamls, d)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	releaseResources, err := r.yamlToAppResource(r.releaseYamls, d)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	postreleaseResources, err := r.yamlToAppResource(r.postreleaseYamls, d)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var workloads []appResource
	var configs []appResource

	for _, releaseResource := range releaseResources {
		if releaseResource.isWorkload() {
			workloads = append(workloads, releaseResource)
		} else if releaseResource.canBeManaged() {
			configs = append(configs, releaseResource)
		}
	}

	var filteredPrereleasers []appResource
	for _, prereleaser := range prereleaseResources {
		if prereleaser.canBeManaged() {
			filteredPrereleasers = append(filteredPrereleasers, prereleaser)
		}
	}
	var filteredPostreleasers []appResource
	for _, postreleaser := range postreleaseResources {
		if postreleaser.canBeManaged() {
			filteredPostreleasers = append(filteredPostreleasers, postreleaser)
		}
	}

	return filteredPrereleasers, configs, workloads, filteredPostreleasers, nil
}

func (r releaser) yamlToAppResource(yamls []string, data map[string]string) (appResources, error) {
	var interpolated [][]byte
	for _, yaml := range yamls {
		i, err := interpolate(yaml, data)
		split := strings.Split(string(i), "\n---\n")
		for _, s := range split {
			interpolated = append(interpolated, []byte(s))
		}
		if err != nil {
			return nil, ErrorContext{err: err, context: "interpolation"}
		}
	}

	var resources appResources
	for _, resourceYaml := range interpolated {
		var parsed parsedResource
		err := yaml.Unmarshal(resourceYaml, &parsed)
		if err != nil {
			return nil, ErrorContext{err: err, context: "unmarshalling raw resources for apply"}
		}

		scope := r.errorScope.AddScope(report.Scope{"resourceName": parsed.Metadata.Name, "resourceKind": parsed.Kind})
		logger := r.logger.With(zap.String("resourceName", parsed.Metadata.Name), zap.String("resourceKind", parsed.Kind))

		var timeout time.Duration
		if t, ok := parsed.Metadata.Annotations["tuber/rolloutTimeout"].(string); ok && t != "" {
			duration, parseErr := time.ParseDuration(t)
			if parseErr != nil {
				return nil, ErrorContext{err: parseErr, context: "invalid timeout", scope: scope, logger: logger}
			}
			timeout = duration
		}

		var rollbackTimeout time.Duration
		if t, ok := parsed.Metadata.Annotations["tuber/rollbackTimeout"].(string); ok && t != "" {
			duration, parseErr := time.ParseDuration(t)
			if parseErr != nil {
				return nil, ErrorContext{err: parseErr, context: "invalid rollback timeout", scope: scope, logger: logger}
			}
			rollbackTimeout = duration
		}

		var watchUrl string
		if t, ok := parsed.Metadata.Annotations["tuber/watchUrl"].(string); ok && t != "" {
			watchUrl = t
		}

		var watchDuration time.Duration
		if t, ok := parsed.Metadata.Annotations["tuber/watchDuration"].(string); ok && t != "" {
			duration, parseErr := time.ParseDuration(t)
			if parseErr != nil {
				return nil, ErrorContext{err: parseErr, context: "invalid watch duration", scope: scope, logger: logger}
			}
			watchDuration = duration
		}

		resource := appResource{
			kind:            parsed.Kind,
			name:            parsed.Metadata.Name,
			contents:        resourceYaml,
			timeout:         timeout,
			rollbackTimeout: rollbackTimeout,
			watchUrl:        watchUrl,
			watchDuration:   watchDuration,
		}

		resources = append(resources, resource)
	}
	return resources, nil
}

func (r releaser) apply(resources []appResource) ([]appResource, error) {
	var applied []appResource
	for _, resource := range resources {
		scope, logger := resource.scopes(r)
		err := k8s.Apply(resource.contents, r.app.Name)
		if err != nil {
			return applied, ErrorContext{err: err, scope: scope, logger: logger, context: "apply"}
		}
		applied = append(applied, resource)
	}
	return applied, nil
}

type rolloutError struct {
	err      error
	resource appResource
}

func (r releaser) watchWorkloads(appliedWorkloads []appResource) error {
	var wg sync.WaitGroup
	errors := make(chan rolloutError)
	done := make(chan bool)
	for _, workload := range appliedWorkloads {
		wg.Add(1)
		var timeout time.Duration
		if workload.timeout == 0 {
			timeout = 5 * time.Minute
		}
		go r.goWatch(workload, timeout, errors, &wg)
	}
	go goWait(&wg, done)
	select {
	case <-done:
		return nil
	case err := <-errors:
		scope, logger := err.resource.scopes(r)
		return ErrorContext{err: err.err, scope: scope, logger: logger, context: "watch workload"}
	}
}

func (r releaser) watchRollback(appliedWorkloads []appResource) []error {
	var wg sync.WaitGroup
	errorChan := make(chan rolloutError)
	done := make(chan bool)
	for _, workload := range appliedWorkloads {
		wg.Add(1)
		var timeout time.Duration
		if workload.rollbackTimeout == 0 {
			timeout = 5 * time.Minute
		}
		go r.goWatchRollback(workload, timeout, errorChan, &wg)
	}
	var errors []error

	go goWait(&wg, done)
	for range appliedWorkloads {
		select {
		case <-done:
			return errors
		case err := <-errorChan:
			scope, logger := err.resource.scopes(r)
			errors = append(errors, ErrorContext{err: err.err, scope: scope, logger: logger, context: "watch rollback"})
		}
	}

	return errors
}

func goWait(wg *sync.WaitGroup, done chan bool) {
	wg.Wait()
	done <- true
}

func (r releaser) goWatch(resource appResource, timeout time.Duration, errors chan rolloutError, wg *sync.WaitGroup) {
	defer wg.Done()
	if !resource.supportsRollback() {
		return
	}

	if resource.watchUrl == "" {
		err := k8s.RolloutStatus(resource.kind, resource.name, r.app.Name, timeout)
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
		}
	} else {
		wg.Add(1)
		go func(errors chan rolloutError, wg *sync.WaitGroup) {
			err := k8s.RolloutStatus(resource.kind, resource.name, r.app.Name, timeout)
			if err != nil {
				errors <- rolloutError{err: err, resource: resource}
			}
			wg.Done()
		}(errors, wg)
		err := r.monitorUrl(resource)
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
		}
	}
}

func (r releaser) monitorUrl(resource appResource) error {
	timeout := time.Now().Add(resource.watchDuration)
	r.logger.Debug("pinging " + resource.watchUrl)
	for {
		if time.Now().After(timeout) {
			return nil
		}
		resp, err := http.Get(resource.watchUrl)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("non-success status received from monitor url")
		}
		time.Sleep(30 * time.Second)
	}
}

func (r releaser) goWatchRollback(resource appResource, timeout time.Duration, errors chan rolloutError, wg *sync.WaitGroup) {
	defer wg.Done()
	if !resource.supportsRollback() {
		return
	}
	err := k8s.RolloutStatus(resource.kind, resource.name, r.app.Name, timeout)
	if err != nil {
		errors <- rolloutError{err: err, resource: resource}
	}
}

func (r releaser) rollback(appliedResources []appResource, cachedResources []appResource) ([]appResource, []error) {
	var rolledBack []appResource
	var errors []error
	var emptyState = len(cachedResources) == 0
	for _, applied := range appliedResources {
		var inPreviousState bool
		scope, logger := applied.scopes(r)

		for _, cached := range cachedResources {
			if applied.kind == cached.kind && applied.name == cached.name {
				inPreviousState = true
				err := r.rollbackResource(applied, cached)
				if err != nil {
					errors = append(errors, r.releaseError(ErrorContext{err: err, context: "rollback", scope: scope, logger: logger}))
					break
				}
				rolledBack = append(rolledBack, applied)
				break
			}
		}

		if !inPreviousState && !emptyState {
			err := k8s.Delete(applied.kind, applied.name, r.app.Name)
			if err != nil {
				errors = append(errors, r.releaseError(ErrorContext{err: err, context: "deleting newly created resource on error", scope: scope, logger: logger}))
			}
		}
	}
	return rolledBack, errors
}

func (r releaser) rollbackResource(applied appResource, cached appResource) error {
	var err error
	if applied.supportsRollback() {
		err = k8s.RolloutUndo(applied.kind, applied.name, r.app.Name)
	} else {
		err = k8s.Apply(cached.contents, r.app.Name)
	}

	if err != nil {
		return err
	}
	return nil
}

func (r releaser) reconcileState(stateBeforeApply []appResource, appliedWorkloads []appResource, appliedConfigs []appResource, appliedPostreleaseResources []appResource) error {
	var appliedResources appResources = append(appliedWorkloads, appliedConfigs...)
	appliedResources = append(appliedResources, appliedPostreleaseResources...)

	type stateResource struct {
		Metadata struct {
			OwnerReferences []map[string]interface{}
		}
	}

	for _, cached := range stateBeforeApply {
		var inPreviousState bool
		for _, applied := range appliedResources {
			if applied.kind == cached.kind && applied.name == cached.name {
				inPreviousState = true
				break
			}
		}
		if !inPreviousState {
			scope, logger := cached.scopes(r)
			out, err := k8s.Get(cached.kind, cached.name, r.app.Name, "-o", "yaml")
			if err != nil {
				if _, ok := err.(k8s.NotFoundError); !ok {
					return ErrorContext{err: err, context: "exists check resource removed from state", scope: scope, logger: logger}
				}
			}

			var parsed stateResource
			err = yaml.Unmarshal(out, &parsed)
			if err != nil {
				return ErrorContext{err: err, context: "parse resource removed from state", scope: scope, logger: logger}
			}

			if parsed.Metadata.OwnerReferences == nil {
				deleteErr := k8s.Delete(cached.kind, cached.name, r.app.Name)
				if deleteErr != nil {
					return ErrorContext{err: err, context: "delete resource removed from state", scope: scope, logger: logger}
				}
			}
		}
	}

	updatedApp := r.app
	updatedApp.State.Previous = r.app.State.Current
	updatedApp.State.Current = appliedResources.encode()

	err := r.db.SaveApp(updatedApp)
	if err != nil {
		return ErrorContext{err: err, context: "save new state"}
	}
	return nil
}

package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/freshly/tuber/pkg/containers"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/report"

	"github.com/goccy/go-yaml"
	"go.uber.org/zap"
)

type releaser struct {
	logger           *zap.Logger
	errorScope       report.Scope
	app              *TuberApp
	digest           string
	data             *ClusterData
	releaseYamls     []string
	prereleaseYamls  []string
	postreleaseYamls []string
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
func Release(yamls containers.AppYamls, logger *zap.Logger, errorScope report.Scope, app *TuberApp, digest string, data *ClusterData) error {
	return releaser{
		logger:           logger,
		errorScope:       errorScope,
		releaseYamls:     yamls.Release,
		prereleaseYamls:  yamls.Prerelease,
		postreleaseYamls: yamls.PostRelease,
		app:              app,
		digest:           digest,
		data:             data,
	}.release()
}

func (r releaser) release() error {
	r.logger.Debug(fmt.Sprintf("releaser starting for %s", r.app.Name))

	r.logger.Debug("getting app ")
	rr, err := r.resourcesToApply()
	if err != nil {
		return r.releaseError(err)
	}

	r.logger.Debug("getting current state")
	state, err := r.currentState()
	if err != nil {
		return r.releaseError(err)
	}

	if len(rr.Prerelease) > 0 {
		r.logger.Debug("prerelease starting")

		err = RunPrerelease(rr.Prerelease, r.app)
		if err != nil {
			return ErrorContext{context: "prerelease", err: err}
		}

		r.logger.Debug("prerelease complete")
	}

	r.logger.Debug("attempting to apply configs")
	appliedConfigs, err := r.apply(rr.Configs)
	if err != nil {
		_ = r.releaseError(err)
		r.rollback(appliedConfigs, state.resources)
		return err
	}

	appliedWorkloads, err := r.apply(rr.Workloads)
	if err != nil {
		_ = r.releaseError(err)
		_, configRollbackErrors := r.rollback(appliedConfigs, state.resources)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, state.resources)
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
		_, configRollbackErrors := r.rollback(appliedConfigs, state.resources)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, state.resources)
		for _, rollbackError := range append(configRollbackErrors, workloadRollbackErrors...) {
			_ = r.releaseError(rollbackError)
		}
		watchErrors := r.watchRollback(rolledBackResources)
		for _, watchError := range watchErrors {
			_ = r.releaseError(watchError)
		}
		return err
	}

	appliedPostreleaseResources, err := r.apply(rr.Postrelease)
	if err != nil {
		_ = r.releaseError(err)
		_, configRollbackErrors := r.rollback(appliedConfigs, state.resources)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, state.resources)
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
		_, configRollbackErrors := r.rollback(appliedConfigs, state.resources)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, state.resources)
		rolledBackPostreleaseResources, postreleaseRollbackErrors := r.rollback(appliedPostreleaseResources, state.resources)
		for _, rollbackError := range append(configRollbackErrors, append(workloadRollbackErrors, postreleaseRollbackErrors...)...) {
			_ = r.releaseError(rollbackError)
		}
		watchErrors := r.watchRollback(append(rolledBackResources, rolledBackPostreleaseResources...))
		for _, watchError := range watchErrors {
			_ = r.releaseError(watchError)
		}
		return err
	}

	err = r.reconcileState(state, appliedWorkloads, appliedConfigs)
	if err != nil {
		return r.releaseError(err)
	}

	return nil
}

type state struct {
	resources []appResource
	raw       rawState
	remote    *k8s.ConfigResource
}

type rawState struct {
	Resources     managedResources `json:"resources"`
	PreviousState managedResources `json:"previousState"`
}

type managedResource struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Encoded string `json:"encoded"`
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

func (a appResources) encode() managedResources {
	var encoded []managedResource
	for _, resource := range a {
		m := managedResource{
			Kind:    resource.kind,
			Name:    resource.name,
			Encoded: base64.StdEncoding.EncodeToString(resource.contents),
		}
		encoded = append(encoded, m)
	}
	return encoded
}

type managedResources []managedResource

func (m managedResources) decode() (appResources, error) {
	var resources appResources
	for _, managed := range m {
		contents, err := base64.StdEncoding.DecodeString(managed.Encoded)
		if err != nil {
			return nil, err
		}
		resources = append(resources, appResource{contents: contents, kind: managed.Kind, name: managed.Name})
	}
	return resources, nil
}

func (r releaser) currentState() (*state, error) {
	stateName := "tuber-state-" + r.app.Name

	exists, err := k8s.Exists("configMap", stateName, r.app.Name)
	if err != nil {
		return nil, ErrorContext{err: err, context: "state config exists check"}
	}

	if !exists {
		r.logger.Debug("config map does not exist for app & state. attempting to create new configmap", zap.String("stateName", stateName), zap.String("appName", r.app.Name))
		createErr := k8s.Create(r.app.Name, "configmap", stateName, `--from-literal=state=`)
		if createErr != nil {
			return nil, ErrorContext{err: createErr, context: "state config creation"}
		}
	}

	stateResource, err := k8s.GetConfigResource(stateName, r.app.Name, "ConfigMap")
	if err != nil {
		return nil, ErrorContext{err: err, context: "get state config"}
	}

	rawStateData := stateResource.Data["state"]

	var stateData rawState
	if rawStateData != "" {
		unmarshalErr := json.Unmarshal([]byte(rawStateData), &stateData)
		if unmarshalErr != nil {
			return nil, ErrorContext{err: unmarshalErr, context: "parse state"}
		}
	}

	resources, err := stateData.Resources.decode()
	if err != nil {
		return nil, ErrorContext{err: err, context: "decode state"}
	}
	return &state{resources: resources, raw: stateData, remote: stateResource}, nil
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

type ResourceCollection struct {
	Prerelease  []appResource
	Configs     []appResource
	Workloads   []appResource
	Postrelease []appResource
}

func (r releaser) resourcesToApply() (*ResourceCollection, error) {
	d := releaseData(r.digest, r.app, r.data)

	prereleaseResources, err := r.yamlToAppResource(r.prereleaseYamls, d)
	if err != nil {
		return nil, err
	}

	releaseResources, err := r.yamlToAppResource(r.releaseYamls, d)
	if err != nil {
		return nil, err
	}

	postreleaseResources, err := r.yamlToAppResource(r.postreleaseYamls, d)
	if err != nil {
		return nil, err
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

	return &ResourceCollection{
		Prerelease:  filteredPrereleasers,
		Configs:     configs,
		Workloads:   workloads,
		Postrelease: filteredPostreleasers}, nil
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

		r.logger.Debug("attempting to apply", zap.String("app name", r.app.Name), zap.ByteString("resource", resource.contents))
		err := k8s.Apply(resource.contents, r.app.Name)
		if err != nil {
			r.logger.Debug("failed to apply", zap.String("app name", r.app.Name), zap.ByteString("resource", resource.contents))
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

	if resource.watchUrl == "" || r.app.ReviewApp {
		err := k8s.RolloutStatus(resource.kind, resource.name, r.app.Name, timeout)
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
		}
	} else {
		go func() {
			err := k8s.RolloutStatus(resource.kind, resource.name, r.app.Name, timeout)
			if err != nil {
				errors <- rolloutError{err: err, resource: resource}
			}
		}()
		err := r.monitorUrl(resource)
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
		}
	}
}

func (r releaser) monitorUrl(resource appResource) error {
	timeout := time.Now().Add(resource.watchDuration)
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

func (r releaser) reconcileState(state *state, appliedWorkloads []appResource, appliedConfigs []appResource) error {
	var appliedResources appResources = append(appliedWorkloads, appliedConfigs...)

	type stateResource struct {
		Metadata struct {
			OwnerReferences []map[string]interface{}
		}
	}

	for _, cached := range state.resources {
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

			if parsed.Metadata.OwnerReferences != nil {
				deleteErr := k8s.Delete(cached.kind, cached.name, r.app.Name)
				if deleteErr != nil {
					return ErrorContext{err: err, context: "delete resource removed from state", scope: scope, logger: logger}
				}
			}
		}
	}

	marshalled, err := json.Marshal(rawState{Resources: appliedResources.encode(), PreviousState: state.raw.Resources})
	if err != nil {
		return ErrorContext{err: err, context: "marshal new state"}
	}

	state.remote.Data["state"] = string(marshalled)
	err = state.remote.Save(r.app.Name)
	if err != nil {
		return ErrorContext{err: err, context: "save new state"}
	}
	return nil
}

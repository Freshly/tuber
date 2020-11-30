package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"tuber/pkg/k8s"
	"tuber/pkg/report"

	"github.com/goccy/go-yaml"
	"go.uber.org/zap"
)

type releaser struct {
	logger       *zap.Logger
	errorScope   report.Scope
	app          *TuberApp
	digest       string
	data         *ClusterData
	releaseYamls []string
	releaseID    string
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
	if err == nil {
		return err
	}

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
func Release(logger *zap.Logger, errorScope report.Scope, releaseYamls []string, app *TuberApp, digest string, data *ClusterData, releaseID string) error {
	return releaser{
		logger:       logger,
		errorScope:   errorScope,
		releaseYamls: releaseYamls,
		app:          app,
		digest:       digest,
		data:         data,
		releaseID:    releaseID,
	}.release()
}

func (r releaser) identifyCanaryTargets(workloads []appResource) ([]appResource, error) {
	var canaries []appResource
	for _, resource := range workloads {
		if resource.isCanary() {
			canaries = append(canaries, resource)
		}
	}

	type canaryResource struct {
		Spec struct {
			TargetRef struct {
				Name string `yaml:"name"`
				Kind string `yaml:"kind"`
			} `yaml:"targetRef"`
		} `yaml:"spec"`
	}

	var updatedWorkloads []appResource
	if len(canaries) > 0 {
		for _, canary := range canaries {
			var unmarshalled canaryResource
			var targetResource appResource
			err := yaml.Unmarshal(canary.contents, &unmarshalled)
			if err != nil {
				return nil, err
			}
			targetRef := unmarshalled.Spec.TargetRef
			var matched bool
			var w []appResource
			for _, resource := range workloads {
				if resource.name == canary.name && resource.kind == canary.kind {
					continue
				}
				if resource.name == targetRef.Name && resource.kind == targetRef.Kind {
					matched = true
					resource.isCanaryTarget = true
					targetResource = resource
				}
				w = append(w, resource)
			}
			if !matched {
				return nil, fmt.Errorf("no canary target found for canary %s", canary.name)
			}
			canary.target = &targetResource
			w = append(w, canary)
			updatedWorkloads = w
		}
	}
	return updatedWorkloads, nil
}

func (r releaser) release() error {
	r.logger.Debug("releaser starting")

	workloads, configs, err := r.resourcesToApply()
	if err != nil {
		return r.releaseError(err)
	}

	workloads, err = r.identifyCanaryTargets(workloads)
	if err != nil {
		return r.releaseError(err)
	}

	state, err := r.currentState()
	if err != nil {
		return r.releaseError(err)
	}

	appliedConfigs, err := r.apply(configs)
	if err != nil {
		_ = r.releaseError(err)
		r.rollback(appliedConfigs, state.resources)
		return err
	}

	appliedWorkloads, err := r.apply(workloads)
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

	err = r.reconcileState(state, appliedWorkloads, appliedConfigs)
	if err != nil {
		return r.releaseError(err)
	}

	return nil
}

func (r releaser) waitForPhase(resource appResource, desiredStatus string, timeout time.Duration, errorStatuses ...string) error {
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	deadline := time.Now().Add(timeout)
	scope, logger := resource.scopes(r)

	for {
		if time.Now().After(deadline) {
			return ErrorContext{err: fmt.Errorf("timeout waiting for %s", desiredStatus), context: "waiting for " + desiredStatus, scope: scope, logger: logger}
		}
		out, err := k8s.Get(resource.kind, resource.name, r.app.Name, `-o=jsonpath="{.status.phase}"`)
		if err != nil {
			return ErrorContext{err: err, context: "get while waiting for " + desiredStatus, scope: scope, logger: logger}
		}
		currentStatus := strings.Trim(string(out), `"`)
		if currentStatus == desiredStatus {
			break
		}
		for _, status := range errorStatuses {
			if currentStatus == status {
				return fmt.Errorf("error status found: %s", status)
			}
		}
		time.Sleep(10 * time.Second)
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
	isCanaryTarget  bool
	target          *appResource
}

func (a appResource) isWorkload() bool {
	return a.supportsRollback() || a.kind == "Pod"
}

func (a appResource) supportsRollback() bool {
	return a.kind == "Deployment" || a.kind == "Daemonset" || a.kind == "StatefulSet" || a.isCanary()
}

func (a appResource) isCanary() bool {
	return a.kind == "Canary"
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
	Name   string                 `yaml:"name"`
	Labels map[string]interface{} `yaml:"labels"`
}

type parsedResource struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   metadata `yaml:"metadata"`
}

func (r releaser) resourcesToApply() ([]appResource, []appResource, error) {
	var interpolated [][]byte
	d := releaseData(r.digest, r.app, r.data, r.releaseID)
	for _, yaml := range r.releaseYamls {
		i, err := interpolate(yaml, d)
		split := strings.Split(string(i), "\n---\n")
		for _, s := range split {
			interpolated = append(interpolated, []byte(s))
		}
		if err != nil {
			return nil, nil, ErrorContext{err: err, context: "interpolation"}
		}
	}

	var workloads []appResource
	var configs []appResource

	for _, resourceYaml := range interpolated {
		var parsed parsedResource
		err := yaml.Unmarshal(resourceYaml, &parsed)
		if err != nil {
			return nil, nil, ErrorContext{err: err, context: "unmarshalling raw resources for apply"}
		}

		scope := r.errorScope.AddScope(report.Scope{"resourceName": parsed.Metadata.Name, "resourceKind": parsed.Kind})
		logger := r.logger.With(zap.String("resourceName", parsed.Metadata.Name), zap.String("resourceKind", parsed.Kind))

		var timeout time.Duration
		if t, ok := parsed.Metadata.Labels["tuber/rolloutTimeout"].(string); ok && t != "" {
			duration, parseErr := time.ParseDuration(t)
			if parseErr != nil {
				return nil, nil, ErrorContext{err: parseErr, context: "invalid timeout", scope: scope, logger: logger}
			}
			timeout = duration
		}

		var rollbackTimeout time.Duration
		if t, ok := parsed.Metadata.Labels["tuber/rollbackTimeout"].(string); ok && t != "" {
			duration, parseErr := time.ParseDuration(t)
			if parseErr != nil {
				return nil, nil, ErrorContext{err: parseErr, context: "invalid rollback timeout", scope: scope, logger: logger}
			}
			rollbackTimeout = duration
		}

		resource := appResource{
			kind:            parsed.Kind,
			name:            parsed.Metadata.Name,
			contents:        resourceYaml,
			timeout:         timeout,
			rollbackTimeout: rollbackTimeout,
		}

		if resource.isWorkload() {
			workloads = append(workloads, resource)
		} else if resource.canBeManaged() {
			configs = append(configs, resource)
		}
	}
	return workloads, configs, nil
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

// TODO: add support for watching pods
func (r releaser) goWatch(resource appResource, timeout time.Duration, errors chan rolloutError, wg *sync.WaitGroup) {
	defer wg.Done()
	if !resource.supportsRollback() {
		return
	}

	if resource.isCanary() {
		out, err := k8s.Get(resource.kind, resource.name, r.app.Name, `-o=jsonpath="{.status.phase}"`)
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
			return
		}

		currentStatus := strings.Trim(string(out), `"`)
		if currentStatus == "Initializing" {
			err = r.waitForPhase(resource, "Initialized", resource.timeout, "Failed")
		} else if currentStatus == "Failed" {
			err = r.waitForPhase(resource, "Progressing", resource.timeout)
			if err != nil {
				errors <- rolloutError{err: err, resource: resource}
				return
			}
			err = r.waitForPhase(resource, "Succeeded", resource.timeout, "Failed")
		} else {
			err = r.waitForPhase(resource, "Progressing", resource.timeout, "Failed")
			if err != nil {
				errors <- rolloutError{err: err, resource: resource}
				return
			}
			err = r.waitForPhase(resource, "Succeeded", resource.timeout, "Failed")
		}
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
		}
	} else {
		err := k8s.RolloutStatus(resource.kind, resource.name, r.app.Name, timeout)
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
		}
	}
}

// whatever flagger. whatever.
func (r releaser) goWatchRollback(resource appResource, timeout time.Duration, errors chan rolloutError, wg *sync.WaitGroup) {
	defer wg.Done()
	if !resource.supportsRollback() {
		return
	}

	if resource.isCanary() {
		err := r.rollbackAndWatchCanary(resource)
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
		}
	} else {
		err := k8s.RolloutStatus(resource.kind, resource.name, r.app.Name, timeout)
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
		}
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
	if applied.isCanaryTarget {
		return nil
	}

	if applied.supportsRollback() {
		if applied.isCanary() {
			return nil
		} else {
			return k8s.RolloutUndo(applied.kind, applied.name, r.app.Name)
		}
	} else {
		return k8s.Apply(cached.contents, r.app.Name)
	}
	return nil
}

func (r releaser) rollbackAndWatchCanary(canary appResource) error {
	out, err := k8s.Get(canary.kind, canary.name, r.app.Name, `-o=jsonpath="{.status.phase}"`)
	if err != nil {
		return ErrorContext{err: err, context: "get while rolling back "}
	}
	currentStatus := strings.Trim(string(out), `"`)
	if currentStatus == "Succeeded" {
		// patch doesn't work and i don't know why. try these locally and they work fine, but kubectl doesn't like it at runtime.
		// err := k8s.Patch(canary.kind, canary.name, r.app.Name, `'[{"op": "add", "path": "/spec/skipAnalysis", "value":true}]'`, "--type=json")
		// err := k8s.Patch(canary.kind, canary.name, r.app.Name, "'{\"spec\": {\"skipAnalysis\": true}}'", "--type=merge")
		// so instead we're doing this shit
		ugh := r.app
		ugh.ReviewApp = true
		d := releaseData(r.digest, ugh, r.data, r.releaseID)
		var reinterpolatedCanary []byte
		for _, yam := range r.releaseYamls {
			i, err := interpolate(yam, d)
			if err != nil {
				return err
			}
			split := strings.Split(string(i), "\n---\n")
			for _, s := range split {
				var parsed parsedResource
				err := yaml.Unmarshal([]byte(s), &parsed)
				if err != nil {
					return err
				}
				if parsed.Metadata.Name == canary.name && parsed.Kind == canary.kind {
					reinterpolatedCanary = []byte(s)
					break
				}
			}
			if string(reinterpolatedCanary) != "" {
				break
			}
		}
		if string(reinterpolatedCanary) == "" {
			return fmt.Errorf("flagger workaround reinterpolated canary fail")
		}

		err := k8s.Apply(reinterpolatedCanary, r.app.Name)
		if err != nil {
			return err
		}

		err = k8s.RolloutUndo(canary.target.kind, canary.target.name, r.app.Name)
		if err != nil {
			return err
		}
	} else if currentStatus != "Failed" {
		err := k8s.PatchConfigMap("flagpole", "flagpole", r.app.Name, "true")
		if err != nil {
			return err
		}

		err = r.waitForPhase(canary, "Failed", canary.rollbackTimeout)
		if err != nil {
			return err
		}

		err = k8s.RemoveConfigMapEntry("flagpole", "flagpole", r.app.Name)
		if err != nil {
			return err
		}
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

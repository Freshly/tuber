package core

import (
	"encoding/base64"
	"encoding/json"
	"sync"
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

	logger.Warn("release error", zap.Error(err), zap.String("context", context))
	report.Error(err, scope)

	return err
}

var nonGenericMetadata = []string{"annotations", "creationTimestamp", "namespace", "resourceVersion", "selfLink", "uid"}

// Deploy interpolates and applies an app's resources. It removes deleted resources, and rolls back on any release failure.
// If you edit a resource manually, and a deploy fails, tuber will roll back to the previously deployed state of the object, not to the state you manually specified.
func Deploy(logger *zap.Logger, errorScope report.Scope, releaseYamls []string, app *TuberApp, digest string, data *ClusterData) error {
	return releaser{
		logger:       logger,
		errorScope:   errorScope,
		releaseYamls: releaseYamls,
		app:          app,
		digest:       digest,
		data:         data,
	}.deploy()
}

func (r releaser) deploy() error {
	r.logger.Debug("releaser starting")

	workloads, configs, err := r.resourcesToApply()
	if err != nil {
		return r.releaseError(err)
	}

	cachedResources, stateResource, err := r.currentState()
	if err != nil {
		return r.releaseError(err)
	}

	appliedConfigs, err := r.apply(configs)
	if err != nil {
		_ = r.releaseError(err)
		r.rollback(appliedConfigs, cachedResources)
		return err
	}

	appliedWorkloads, err := r.apply(workloads)
	if err != nil {
		_ = r.releaseError(err)
		_, configRollbackErrors := r.rollback(appliedConfigs, cachedResources)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, cachedResources)
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
		_, configRollbackErrors := r.rollback(appliedConfigs, cachedResources)
		rolledBackResources, workloadRollbackErrors := r.rollback(appliedWorkloads, cachedResources)
		for _, rollbackError := range append(configRollbackErrors, workloadRollbackErrors...) {
			_ = r.releaseError(rollbackError)
		}
		watchErrors := r.watchRollback(rolledBackResources)
		for _, watchError := range watchErrors {
			_ = r.releaseError(watchError)
		}
		return err
	}

	err = r.reconcileState(appliedWorkloads, appliedConfigs, cachedResources, stateResource)
	if err != nil {
		return r.releaseError(err)
	}

	return nil
}

type appState struct {
	Resources []managedResource `json:"resources"`
}

type managedResource struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Encoded string `json:"encoded"`
}

type appResource struct {
	contents []byte
	kind     string
	name     string
}

func (a appResource) isWorkload() bool {
	return a.supportsRollback() || a.kind == "Pod"
}

func (a appResource) supportsRollback() bool {
	return a.kind == "Deployment" || a.kind == "Daemonset" || a.kind == "StatefulSet" || a.isRollout()
}

func (a appResource) isRollout() bool {
	return a.kind == "Rollout"
}

func (a appResource) canBeManaged() bool {
	return a.kind != "Secret" && a.kind != "Role" && a.kind != "RoleBinding" && a.kind != "ClusterRole" && a.kind != "ClusterRoleBinding"
}

func (a appResource) scopes(r releaser) (report.Scope, *zap.Logger) {
	scope := r.errorScope.AddScope(report.Scope{"resourceName": a.name, "resourceKind": a.kind})
	logger := r.logger.With(zap.String("resourceName", a.name), zap.String("resourceKind", a.kind))
	return scope, logger
}

func (r releaser) currentState() ([]appResource, *k8s.ConfigResource, error) {
	stateName := "tuber-state-" + r.app.Name
	exists, err := k8s.Exists("configMap", stateName, r.app.Name)

	if err != nil {
		return nil, nil, ErrorContext{err: err, context: "state config exists check"}
	}

	if !exists {
		err := k8s.Create(r.app.Name, "configmap", stateName, `--from-literal=state=`)
		if err != nil {
			return nil, nil, ErrorContext{err: err, context: "state config creation"}
		}
	}

	stateResource, err := k8s.GetConfigResource(stateName, r.app.Name, "ConfigMap")
	if err != nil {
		return nil, nil, ErrorContext{err: err, context: "get state config"}
	}

	rawState := stateResource.Data["state"]

	var state *appState
	if rawState != "" {
		jsonErr := json.Unmarshal([]byte(rawState), &state)
		if jsonErr != nil {
			return nil, nil, ErrorContext{err: err, context: "parse state"}
		}
	}

	var genericized []appResource
	for _, resource := range state.Resources {
		contents, decodeErr := base64.StdEncoding.DecodeString(resource.Encoded)
		if decodeErr != nil {
			return nil, nil, ErrorContext{err: err, context: "decode contents"}
		}
		genericized = append(genericized, appResource{contents: contents, kind: resource.Kind, name: resource.Name})
	}
	return genericized, stateResource, nil
}

type metadata struct {
	Name   string
	Labels map[string]string
}

type parsedResource struct {
	ApiVersion string
	Kind       string
	Metadata   metadata
}

func (r releaser) resourcesToApply() ([]appResource, []appResource, error) {
	var interpolated [][]byte
	d := releaseData(r.digest, r.app, r.data)
	for _, yaml := range r.releaseYamls {
		i, err := interpolate(yaml, d)
		interpolated = append(interpolated, i)
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
			return nil, nil, ErrorContext{err: err, context: "inspecting raw resources for apply"}
		}

		resource := appResource{
			kind:     parsed.Kind,
			name:     parsed.Metadata.Name,
			contents: resourceYaml,
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
		var err error
		scope, logger := resource.scopes(r)
		err = k8s.Apply(resource.contents, r.app.Name)
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
		go r.goWatch(workload, errors, &wg)
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
		go r.goWatch(workload, errorChan, &wg)
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
// TODO: add support for watching argo rollouts
func (r releaser) goWatch(resource appResource, errors chan rolloutError, wg *sync.WaitGroup) {
	if resource.supportsRollback() && !resource.isRollout() {
		err := k8s.RolloutStatus(resource.kind, resource.name, r.app.Name)
		if err != nil {
			errors <- rolloutError{err: err, resource: resource}
		}
	}
	wg.Done()
}

func (r releaser) rollback(appliedResources []appResource, cachedResources []appResource) ([]appResource, []error) {
	var rolledBack []appResource
	var errors []error
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
		if !inPreviousState {
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
		if applied.isRollout() {
			// TODO: add actual argo support
			err = k8s.Apply(cached.contents, r.app.Name)
		} else {
			err = k8s.RolloutUndo(applied.kind, applied.name, r.app.Name)
		}
	} else {
		err = k8s.Apply(cached.contents, r.app.Name)
	}

	if err != nil {
		return err
	}
	return nil
}

func (r releaser) reconcileState(appliedWorkloads []appResource, appliedConfigs []appResource, cachedResources []appResource, stateResource *k8s.ConfigResource) error {
	appliedResources := append(appliedWorkloads, appliedConfigs...)
	for _, cached := range cachedResources {
		var inPreviousState bool
		for _, applied := range appliedResources {
			if applied.kind == cached.kind && applied.name == cached.name {
				inPreviousState = true
				break
			}
		}
		if !inPreviousState {
			scope, logger := cached.scopes(r)
			err := k8s.Delete(cached.kind, cached.name, r.app.Name)
			if err != nil {
				return ErrorContext{err: err, context: "delete resources removed from state", scope: scope, logger: logger}
			}
		}
	}

	var updatedState []managedResource
	for _, resource := range appliedResources {
		stateResource := managedResource{
			Kind:    resource.kind,
			Name:    resource.name,
			Encoded: base64.StdEncoding.EncodeToString(resource.contents),
		}
		updatedState = append(updatedState, stateResource)
	}

	marshalled, err := json.Marshal(appState{Resources: updatedState})
	if err != nil {
		return ErrorContext{err: err, context: "marshal new state"}
	}

	stateResource.Data["state"] = string(marshalled)
	err = stateResource.Save(r.app.Name)
	if err != nil {
		return ErrorContext{err: err, context: "save new state"}
	}
	return nil
}

// deprecated and unused, but a hopefully useful example of resource editing
// func addAnnotationToV1Deployment(resource []byte) (string, string, error) {
// 	decode := scheme.Codecs.UniversalDeserializer().Decode
//
// 	obj, versionKind, err := decode(resource, nil, nil)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	if versionKind.Version != "v1" {
// 		return "", "", fmt.Errorf("must use v1 deployments")
// 	}
//
// 	deployment := obj.(*v1.Deployment)
// 	annotations := deployment.Spec.Template.ObjectMeta.GetAnnotations()
// 	if annotations == nil {
// 		annotations = map[string]string{}
// 	}
// 	releaseID := uuid.New().String()
// 	annotations["tuber/releaseID"] = releaseID
// 	deployment.Spec.Template.ObjectMeta.SetAnnotations(annotations)
//
// 	annotated, err := yaml.Marshal(deployment)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	return string(annotated), releaseID, nil
// }

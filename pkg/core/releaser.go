package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"tuber/pkg/k8s"
	"tuber/pkg/report"

	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type releaser struct {
	logger     *zap.Logger
	errorScope report.Scope
	app        *TuberApp
	digest     string
	data       *ClusterData
	tubers     []string
}

type ErrorContext struct {
	scope   report.Scope
	logger  *zap.Logger
	err     error
	context string
}

func (r releaser) releaseError(err error) error {
	var context string
	var scope = r.errorScope
	var logger = r.logger

	errorContext, ok := err.(*ErrorContext)
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

	logger.Warn("failed release", zap.Error(err), zap.String("context", context))
	report.Error(err, scope)

	return err
}

func (e ErrorContext) Error() string {
	return e.err.Error()
}

type appState struct {
	Resources []appResource `json:"resources"`
}

type appResource struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Encoded string `json:"encoded"`
}

type managedResource struct {
	contents []byte
	kind     string
	name     string
}

var nonGenericMetadata = []string{"annotations", "creationTimestamp", "namespace", "resourceVersion", "selfLink", "uid"}

// ReleaseTubers interpolates and applies an app's resources. It removes deleted resources, and rolls back on any release failure.
// If you edit a resource manually, and a deploy fails, tuber will roll back to the previously tuberized state of the object, not to the state you manually specified.
func ReleaseTubers(logger *zap.Logger, errorScope report.Scope, tubers []string, app *TuberApp, digest string, data *ClusterData) {
	releaser{
		logger:     logger,
		errorScope: errorScope,
		tubers:     tubers,
		app:        app,
		digest:     digest,
		data:       data,
	}.releaseTubers()
}

func (r releaser) releaseTubers() error {
	r.logger.Debug("release starting")
	startTime := time.Now()

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
		r.releaseError(err)
		r.rollback(appliedConfigs, cachedResources)
		return err
	}

	appliedWorkloads, err := r.apply(workloads)
	if err != nil {
		r.releaseError(err)
		r.rollback(appliedConfigs, cachedResources)
		r.rollback(appliedWorkloads, cachedResources)
		return err
	}

	err = r.watch(appliedWorkloads)
	if err != nil {
		r.releaseError(err)
		r.rollback(appliedConfigs, cachedResources)
		r.rollback(appliedWorkloads, cachedResources)
		return err
	}

	err = r.reconcileState(appliedWorkloads, appliedConfigs, cachedResources, stateResource)
	if err != nil {
		return r.releaseError(err)
	}

	r.logger.Info("release complete", zap.Duration("duration", time.Since(startTime)))
	return nil
}

func (r releaser) currentState() ([]managedResource, *k8s.ConfigResource, error) {
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

	var genericized []managedResource
	for _, resource := range state.Resources {
		contents, decodeErr := base64.StdEncoding.DecodeString(resource.Encoded)
		if decodeErr != nil {
			return nil, nil, ErrorContext{err: err, context: "decode contents"}
		}
		genericized = append(genericized, managedResource{contents: contents, kind: resource.Kind, name: resource.Name})
	}
	return genericized, stateResource, nil
}

func (r releaser) resourcesToApply() ([]k8sResource, []k8sResource, error) {
	var interpolated []string
	d := tuberData(r.digest, r.app, r.data)
	for _, tuber := range r.tubers {
		i, err := interpolate(tuber, d)
		interpolated = append(interpolated, string(i))
		if err != nil {
			return nil, nil, ErrorContext{err: err, context: "interpolation"}
		}
	}

	var workloads []k8sResource
	var configs []k8sResource

	for _, resourceYaml := range interpolated {
		var resource k8sResource
		err := yaml.Unmarshal([]byte(resourceYaml), &resource)
		if err != nil {
			return nil, nil, ErrorContext{err: err, context: "inspecting raw resources for apply"}
		}
		resource.raw = resourceYaml

		if resource.isWorkload() {
			workloads = append(workloads, resource)
		} else if resource.canBeManaged() {
			configs = append(configs, resource)
		}
	}
	return workloads, configs, nil
}

func (r releaser) apply(resources []k8sResource) ([]k8sResource, error) {
	var applied []k8sResource
	for _, resource := range resources {
		var err error
		scope := r.errorScope.AddScope(report.Scope{"resourceName": resource.Metadata.Name, "resourceKind": resource.Kind})
		logger := r.logger.With(zap.String("resourceName", resource.Metadata.Name), zap.String("resourceKind", resource.Kind))
		err = k8s.Apply([]byte(resource.raw), r.app.Name)
		if err != nil {
			return applied, ErrorContext{err: err, scope: scope, logger: logger, context: "apply"}
		}
		applied = append(applied, resource)
	}
	return applied, nil
}

func (r releaser) watch(appliedWorkloads []k8sResource) error {
	var wg sync.WaitGroup
	errors := make(chan error)
	done := make(chan bool, len(appliedWorkloads))
	for _, workload := range appliedWorkloads {
		wg.Add(1)
		go r.goWatch(workload, errors, &wg)
	}
	go goWait(&wg, done)
	select {
	case <-done:
		return nil
	case err := <-errors:
		return err
	}
}

func goWait(wg *sync.WaitGroup, done chan bool) {
	wg.Wait()
	done <- true
}

// TODO: add support for watching pods
// TODO: add support for watching argo rollouts
func (r releaser) goWatch(resource k8sResource, errors chan error, wg *sync.WaitGroup) {
	if resource.supportsRollback() && !resource.isRollout() {
		err := k8s.RolloutStatus(resource.Kind, resource.Metadata.Name, r.app.Name)
		if err != nil {
			errors <- err
		}
	}
	wg.Done()
}

func (r releaser) rollback(appliedResources []k8sResource, cachedResources []managedResource) {
	for _, applied := range appliedResources {
		var inPreviousState bool
		scope := r.errorScope.AddScope(report.Scope{"resourceName": applied.Metadata.Name, "resourceKind": applied.Kind})
		logger := r.logger.With(zap.String("resourceName", applied.Metadata.Name), zap.String("resourceKind", applied.Kind))

		for _, cached := range cachedResources {
			if applied.Kind == cached.kind && applied.Metadata.Name == cached.name {
				inPreviousState = true
				err := r.rollbackResource(applied, cached)
				if err != nil {
					_ = r.releaseError(ErrorContext{err: err, context: "rollback", scope: scope, logger: logger})
					break
				}
				break
			}
		}
		if !inPreviousState {
			err := k8s.Delete(applied.Kind, applied.Metadata.Name, r.app.Name)
			if err != nil {
				_ = r.releaseError(ErrorContext{err: err, context: "deleting newly created resource on error", scope: scope, logger: logger})
			}
		}
	}
	return
}

func (r releaser) rollbackResource(applied k8sResource, cached managedResource) error {
	err := k8s.Apply(cached.contents, r.app.Name)
	if err != nil {
		return err
	}
	return nil
}

func (r releaser) reconcileState(appliedWorkloads []k8sResource, appliedConfigs []k8sResource, cachedResources []managedResource, stateResource *k8s.ConfigResource) error {
	appliedResources := append(appliedWorkloads, appliedConfigs...)
	for _, cached := range cachedResources {
		var inPreviousState bool
		for _, applied := range appliedResources {
			if applied.Kind == cached.kind && applied.Metadata.Name == cached.name {
				inPreviousState = true
				break
			}
		}
		if !inPreviousState {
			scope := r.errorScope.AddScope(report.Scope{"resourceName": cached.name, "resourceKind": cached.kind})
			logger := r.logger.With(zap.String("resourceName", cached.name), zap.String("resourceKind", cached.kind))
			err := k8s.Delete(cached.kind, cached.name, r.app.Name)
			if err != nil {
				return ErrorContext{err: err, context: "delete resources removed from state", scope: scope, logger: logger}
			}
		}
	}

	var appliedTuberResources []appResource
	for _, resource := range appliedResources {
		stateResource := appResource{
			Kind:    resource.Kind,
			Name:    resource.Metadata.Name,
			Encoded: base64.StdEncoding.EncodeToString([]byte(resource.raw)),
		}
		appliedTuberResources = append(appliedTuberResources, stateResource)
	}

	marshalled, err := json.Marshal(appState{Resources: appliedTuberResources})
	if err != nil {
		return ErrorContext{err: err, context: "marshal new tuber state"}
	}

	stateResource.Data["state"] = string(marshalled)
	err = stateResource.Save(r.app.Name)
	if err != nil {
		return ErrorContext{err: err, context: "save new tuber state"}
	}
	return nil
}

// ClusterData is configurable, cluster-wide data available for yaml interpolation
type ClusterData struct {
	DefaultGateway string
	DefaultHost    string
}

func tuberData(digest string, app *TuberApp, clusterData *ClusterData) (data map[string]string) {
	return map[string]string{
		"tuberImage":            digest,
		"clusterDefaultGateway": clusterData.DefaultGateway,
		"clusterDefaultHost":    clusterData.DefaultHost,
		"tuberAppName":          app.Name,
	}
}

type k8sResource struct {
	ApiVersion string
	Kind       string
	Metadata   metadata
	raw        string
}

type metadata struct {
	Name   string
	Labels map[string]string
}

func (r k8sResource) isWorkload() bool {
	return r.supportsRollback() || r.Kind == "Pod"
}

func (r k8sResource) supportsRollback() bool {
	return r.Kind == "Deployment" || r.Kind == "Daemonset" || r.Kind == "StatefulSet" || r.isRollout()
}

func (r k8sResource) isRollout() bool {
	return r.Kind == "Rollout"
}

func (r k8sResource) canBeManaged() bool {
	return r.Kind != "Secret" && r.Kind != "Role" && r.Kind != "RoleBinding" && r.Kind != "ClusterRole" && r.Kind != "ClusterRoleBinding"
}

// deprecated and unused, but a hopefully useful example of resource editing
func addAnnotationToV1Deployment(resource []byte) (string, string, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, versionKind, err := decode(resource, nil, nil)
	if err != nil {
		return "", "", err
	}
	if versionKind.Version != "v1" {
		return "", "", fmt.Errorf("must use v1 deployments")
	}

	deployment := obj.(*v1.Deployment)
	annotations := deployment.Spec.Template.ObjectMeta.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	releaseID := uuid.New().String()
	annotations["tuber/releaseID"] = releaseID
	deployment.Spec.Template.ObjectMeta.SetAnnotations(annotations)

	annotated, err := yaml.Marshal(deployment)
	if err != nil {
		return "", "", err
	}
	return string(annotated), releaseID, nil
}

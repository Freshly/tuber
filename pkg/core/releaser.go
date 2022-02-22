package core

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/gcr"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/monitor"
	"github.com/freshly/tuber/pkg/report"
	"github.com/freshly/tuber/pkg/slack"

	"github.com/goccy/go-yaml"
	"go.uber.org/zap"
)

type releaser struct {
	logger            *zap.Logger
	errorScope        report.Scope
	app               *model.TuberApp
	digest            string
	data              *ClusterData
	releaseYamls      []string
	prereleaseYamls   []string
	postreleaseYamls  []string
	tags              []string
	db                *DB
	slackClient       *slack.Client
	diffText          string
	sentryBearerToken string
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

	if err == nil {
		return nil
	}

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
	scope.AddScope(report.Scope{"tags": strings.Join(r.tags, ",")})

	logger.Error("release error", zap.Error(err), zap.String("context", context))
	report.Error(err, scope)

	return err
}

// Release interpolates and applies an app's resources. It removes deleted resources, and rolls back on any release failure.
// If you edit a resource manually, and a release fails, tuber will roll back to the previously released state of the object, not to the state you manually specified.
func Release(db *DB, yamls *gcr.AppYamls, logger *zap.Logger, errorScope report.Scope, app *model.TuberApp, digest string, data *ClusterData, slackClient *slack.Client, diffText string, sentryBearerToken string) error {
	return releaser{
		logger:            logger,
		errorScope:        errorScope,
		releaseYamls:      yamls.Release,
		prereleaseYamls:   yamls.Prerelease,
		postreleaseYamls:  yamls.PostRelease,
		tags:              yamls.Tags,
		app:               app,
		digest:            digest,
		data:              data,
		db:                db,
		slackClient:       slackClient,
		diffText:          diffText,
		sentryBearerToken: sentryBearerToken,
	}.release()
}

func (r releaser) release() error {
	r.logger.Debug("releaser starting")
	r.slackClient.Message(r.logger, ":game_die: *"+r.app.Name+"*: release starting"+r.diffText, r.app.SlackChannel)

	rr, err := r.resourcesToApply()
	if err != nil {
		return r.releaseError(err)
	}

	decodedStateBeforeApply, err := r.currentState()
	if err != nil {
		return r.releaseError(err)
	}

	// magnificent
	crtg := currentReplicasToGather(append(rr.Prerelease, append(rr.Workloads, append(rr.Configs, rr.Postrelease...)...)...))

	if len(rr.Prerelease) > 0 {
		r.logger.Debug("prerelease starting")

		err = RunPrerelease(r.logger, r.applyCurrentReplicasToCollection(rr.Prerelease, crtg), r.app)
		if err != nil {
			return ErrorContext{context: "prerelease", err: err}
		}

		r.logger.Debug("prerelease complete")
	}

	appliedConfigs, err := r.apply(rr.Configs)
	if err != nil {
		_ = r.releaseError(err)
		r.rollback(appliedConfigs, decodedStateBeforeApply)
		return err
	}

	appliedWorkloads, err := r.apply(r.applyCurrentReplicasToCollection(rr.Workloads, crtg))
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

	if len(rr.Postrelease) != 0 {
		r.slackClient.Message(r.logger, ":bird: *"+r.app.Name+"*: canary rollout starting"+r.diffText, r.app.SlackChannel)
	}

	rolloutErr, err := r.watchWorkloads(appliedWorkloads)
	if err != nil {
		if !rolloutErr.monitorFail {
			_ = r.releaseError(err)
		} else {
			r.slackClient.Message(r.logger, "<!here> :loudspeaker: *"+r.app.Name+"*: monitoring failed for "+strings.ToLower(rolloutErr.resource.kind)+" "+rolloutErr.resource.name+" - "+rolloutErr.monitorFailMessage, r.app.SlackChannel)
			r.app.Paused = true
			saveErr := r.db.SaveApp(r.app)
			if saveErr != nil {
				_ = r.releaseError(ErrorContext{context: "save app as paused on monitor fail", err: saveErr})
			}
		}
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

	if len(rr.Postrelease) != 0 {
		r.slackClient.Message(r.logger, ":bird: *"+r.app.Name+"*: deployed to canary"+r.diffText, r.app.SlackChannel)
	}

	appliedPostreleaseResources, err := r.apply(r.applyCurrentReplicasToCollection(rr.Postrelease, crtg))
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

	rolloutErr, err = r.watchWorkloads(appliedPostreleaseResources)
	if err != nil {
		if !rolloutErr.monitorFail {
			_ = r.releaseError(err)
		} else {
			r.slackClient.Message(r.logger, "<!here> :loudspeaker: *"+r.app.Name+"*: monitoring failed for "+rolloutErr.resource.kind+" "+rolloutErr.resource.name+"+"+rolloutErr.monitorFailMessage, r.app.SlackChannel)
			r.app.Paused = true
			saveErr := r.db.SaveApp(r.app)
			if saveErr != nil {
				_ = r.releaseError(ErrorContext{context: "save app as paused on monitor fail", err: saveErr})
			}
		}
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

	var appliedResources appResources = append(appliedWorkloads, appliedConfigs...)
	appliedResources = append(appliedResources, appliedPostreleaseResources...)

	cleanupErr := r.deleteRemovedResources(decodedStateBeforeApply, appliedResources)
	if cleanupErr != nil {
		r.slackClient.Message(r.logger, "<!here> :confused: *"+r.app.Name+"*: Release is complete, but deletion of a resource removed with this release failed."+r.diffText, r.app.SlackChannel)
	}

	saveStateErr := r.updateState(appliedResources)
	if saveStateErr != nil {
		r.slackClient.Message(r.logger, "<!here> :skull_and_crossbones: *"+r.app.Name+"*: Release is complete, but the current and previous states *failed to update*.\nRolling back from this release is therefore NOT necessarily safe. Please contact devops."+r.diffText, r.app.SlackChannel)
	}
	if cleanupErr != nil {
		return r.releaseError(cleanupErr)
	}
	if saveStateErr != nil {
		return r.releaseError(saveStateErr)
	}

	return nil
}

type appResource struct {
	contents                     []byte
	kind                         string
	name                         string
	timeout                      time.Duration
	rollbackTimeout              time.Duration
	sentryUrls                   []string
	watchDuration                time.Duration
	hpaCurrentReplicasDeployment string
	hpaCurrentReplicasHpa        string
}

func (a appResource) hasMonitoring() bool {
	return a.watchDuration != 0
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

type ResourceCollection struct {
	Prerelease  []appResource
	Configs     []appResource
	Workloads   []appResource
	Postrelease []appResource
}

func exclusionKey(kind string, name string) string {
	return strings.ToLower(kind) + ":" + strings.ToLower(name)
}

func exclusions(app *model.TuberApp, data map[string]string) (map[string]bool, error) {
	exc := make(map[string]bool)
	for _, resource := range app.ExcludedResources {
		name, err := interpolate(resource.Name, data)
		if err != nil {
			return nil, err
		}
		exc[exclusionKey(resource.Kind, string(name))] = true
	}
	return exc, nil
}

func currentReplicasToGather(resources appResources) map[string]string {
	c := make(map[string]string)
	for _, resource := range resources {
		if resource.hpaCurrentReplicasDeployment != "" {
			if resource.hpaCurrentReplicasHpa != "" {
				c[resource.hpaCurrentReplicasDeployment] = resource.hpaCurrentReplicasHpa
				continue
			}
			c[resource.hpaCurrentReplicasDeployment] = resource.name
		}
	}
	return c
}

func (r releaser) applyCurrentReplicas(contents []byte, hpaName string) []byte {
	var d map[interface{}]interface{}
	err := yaml.Unmarshal(contents, &d)
	if err != nil {
		r.logger.Warn(fmt.Sprintf("error unmarshalling for applyCurrentReplicas: %v", err))
		return nil
	}

	spec, ok := d["spec"].(map[string]interface{})
	if !ok {
		r.logger.Warn("error asserting deployment contents as map[string]interface{} for applyCurrentReplicas")
		return nil
	}

	out, err := k8s.Get("hpa", hpaName, r.app.Name, `-o=go-template="{{.status.desiredReplicas}}"`)
	if err != nil {
		r.logger.Warn(fmt.Sprintf("error getting hpa applyCurrentReplicas: %v", err))
		return nil
	}
	stringOut := strings.Trim(string(out), "\n\r\" ")
	count, err := strconv.Atoi(stringOut)
	if err != nil {
		r.logger.Warn(fmt.Sprintf("error converting hpa desired replicas (stringout: %s) to int for applyCurrentReplicas: %v", stringOut, err))
		return nil
	}
	spec["replicas"] = count
	d["spec"] = spec
	out, err = yaml.Marshal(d)
	if err != nil {
		r.logger.Warn(fmt.Sprintf("error remarshalling deployment with count: %v, for applyCurrentReplicas: %v", count, err))
		return nil
	}
	return out
}

func (r releaser) applyCurrentReplicasToCollection(resources appResources, currentReplicasToGather map[string]string) appResources {
	var updated appResources
	for _, resource := range resources {
		if resource.kind == "Deployment" && currentReplicasToGather[resource.name] != "" {
			updatedContents := r.applyCurrentReplicas(resource.contents, currentReplicasToGather[resource.name])
			if updatedContents != nil {
				resource.contents = updatedContents
			}
		}
		updated = append(updated, resource)
	}
	return updated
}

func (r releaser) resourcesToApply() (*ResourceCollection, error) {
	d := releaseData(r.app.ImageTag, r.app, r.data)

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
		} else {
			configs = append(configs, releaseResource)
		}
	}

	return &ResourceCollection{
		Prerelease:  prereleaseResources,
		Configs:     configs,
		Workloads:   workloads,
		Postrelease: postreleaseResources}, nil
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

	exc, err := exclusions(r.app, data)
	if err != nil {
		return nil, ErrorContext{err: err, context: "interpolation"}
	}

	var resources appResources
	for _, resourceYaml := range interpolated {
		var parsed parsedResource
		err = yaml.Unmarshal(resourceYaml, &parsed)
		if err != nil {
			return nil, ErrorContext{err: err, context: "unmarshalling raw resources for apply"}
		}

		if exc[exclusionKey(parsed.Kind, parsed.Metadata.Name)] {
			continue
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

		var hpaCurrentReplicasTargetDeployment string
		var hpaCurrentReplicasHpa string
		lowerKind := strings.ToLower(parsed.Kind)
		if lowerKind == "horizontalpodautoscaler" || lowerKind == "scaledobject" {
			if t, ok := parsed.Metadata.Annotations["tuber/currentReplicasDeployment"].(string); ok && t != "" {
				hpaCurrentReplicasTargetDeployment = t
			}
			if t, ok := parsed.Metadata.Annotations["tuber/currentReplicasHpa"].(string); ok && t != "" {
				hpaCurrentReplicasHpa = t
			}
		}

		var sentryUrls []string
		for k, v := range parsed.Metadata.Annotations {
			if strings.HasPrefix(k, "tuber/sentryUrl") {
				url, ok := v.(string)
				if ok && url != "" {
					sentryUrls = append(sentryUrls, url)
				}
			}
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
			kind:                         parsed.Kind,
			name:                         parsed.Metadata.Name,
			contents:                     resourceYaml,
			timeout:                      timeout,
			rollbackTimeout:              rollbackTimeout,
			sentryUrls:                   sentryUrls,
			watchDuration:                watchDuration,
			hpaCurrentReplicasDeployment: hpaCurrentReplicasTargetDeployment,
			hpaCurrentReplicasHpa:        hpaCurrentReplicasHpa,
		}

		if resource.canBeManaged() {
			resources = append(resources, resource)
		}

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
	err                error
	resource           appResource
	monitorFail        bool
	monitorFailMessage string
}

func (r releaser) watchWorkloads(appliedWorkloads []appResource) (rolloutError, error) {
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
		return rolloutError{}, nil
	case err := <-errors:
		scope, logger := err.resource.scopes(r)
		if err.monitorFail {
			r.logger.Warn("release error", zap.Error(err.err), zap.String("context", "monitor url"))
		}
		return err, ErrorContext{err: err.err, scope: scope, logger: logger, context: "watch workload"}
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

	if r.app.Name == "tuber" {
		return
	}

	if !resource.hasMonitoring() {
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

		for _, url := range resource.sentryUrls {
			wg.Add(1)
			go func(url string, errors chan rolloutError, wg *sync.WaitGroup) {
				_, logger := resource.scopes(r)
				healthy, message := monitor.Sentry(logger, url, r.sentryBearerToken, resource.watchDuration)
				if !healthy {
					errors <- rolloutError{
						err:                fmt.Errorf("sentry monitoring found failure"),
						resource:           resource,
						monitorFail:        true,
						monitorFailMessage: message,
					}
				}
				wg.Done()
			}(url, errors, wg)
		}
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
		err = k8s.Apply(cached.contents, r.app.Name)
	} else {
		err = k8s.Apply(cached.contents, r.app.Name)
	}

	if err != nil {
		return err
	}
	return nil
}

func (r releaser) deleteRemovedResources(stateBeforeApply []appResource, appliedResources appResources) error {
	type stateResource struct {
		Metadata struct {
			OwnerReferences []map[string]interface{}
		}
	}

	for _, cached := range stateBeforeApply {
		var cachedWasPartOfRelease bool
		for _, applied := range appliedResources {
			if applied.kind == cached.kind && applied.name == cached.name {
				cachedWasPartOfRelease = true
				break
			}
		}
		if !cachedWasPartOfRelease {
			scope, logger := cached.scopes(r)
			out, err := k8s.Get(cached.kind, cached.name, r.app.Name, "-o", "yaml")
			if _, notFound := err.(k8s.NotFoundError); notFound {
				continue
			}
			if err != nil {
				return ErrorContext{err: err, context: "exists check resource removed from state", scope: scope, logger: logger}
			}

			var parsed stateResource
			err = yaml.Unmarshal(out, &parsed)
			if err != nil {
				return ErrorContext{err: err, context: "parse resource removed from state", scope: scope, logger: logger}
			}

			if parsed.Metadata.OwnerReferences == nil {
				deleteErr := k8s.Delete(cached.kind, cached.name, r.app.Name)
				if deleteErr != nil {
					return ErrorContext{err: deleteErr, context: "delete resource removed from state", scope: scope, logger: logger}
				}
			}
		}
	}
	return nil
}

func (r releaser) updateState(appliedResources appResources) error {
	latest, err := r.db.App(r.app.Name)
	if err != nil {
		return ErrorContext{err: err, context: "save new state and tags: pull latest app data"}
	}

	latest.State.Previous = latest.State.Current
	latest.State.Current = appliedResources.encode()
	latest.CurrentTags = r.tags

	err = r.db.SaveApp(latest)
	if err != nil {
		return ErrorContext{err: err, context: "save new state and tags"}
	}
	return nil
}

package reviewapps

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/google/go-containerregistry/pkg/name"

	"go.uber.org/zap"
)

// NewReviewAppSetup replicates a namespace and its roles, rolebindings, and opaque secrets after removing their non-generic metadata.
// Also renames source app name matches across all relevant resources.
func NewReviewAppSetup(sourceApp string, reviewApp string) error {
	err := copyNamespace(sourceApp, reviewApp)
	if err != nil {
		return err
	}
	for _, kind := range []string{"roles", "rolebindings"} {
		rolesErr := copyResources(kind, sourceApp, reviewApp)
		if rolesErr != nil {
			return rolesErr
		}
	}
	err = copyResources("secrets", sourceApp, reviewApp, "--field-selector", "type=Opaque")
	if err != nil {
		return err
	}

	return nil
}

func CreateReviewApp(ctx context.Context, db *core.DB, l *zap.Logger, branch string, appName string, credentials []byte, projectName string) (string, error) {
	reviewAppName := ReviewAppName(appName, branch)

	if db.AppExists(reviewAppName) {
		return "", fmt.Errorf("review app already exists")
	}

	logger := l.With(
		zap.String("appName", appName),
		zap.String("reviewAppName", reviewAppName),
		zap.String("branch", branch),
	)

	logger.Info("creating review app")

	sourceApp, err := db.App(appName)
	if err != nil {
		return "", fmt.Errorf("can't find source app. is %s managed by tuber", appName)
	}

	if sourceApp.ReviewAppsConfig == nil || !sourceApp.ReviewAppsConfig.Enabled {
		return "", fmt.Errorf("source app is not enabled for review apps")
	}

	sourceAppTagGCRRef, err := name.ParseReference(sourceApp.ImageTag)
	if err != nil {
		return "", fmt.Errorf("source app image tag misconfigured: %v", err)
	}

	logger.Info("creating review app resources")

	err = NewReviewAppSetup(appName, reviewAppName)
	if err != nil {
		return "", err
	}

	logger.Info("creating and running review app trigger")

	triggerID, err := CreateAndRunTrigger(ctx, logger, credentials, projectName, branch, sourceApp.CloudSourceRepo, reviewAppName)
	if err != nil {
		logger.Error("failed to create or run review app", zap.Error(err))
		triggerCleanupErr := deleteReviewAppTrigger(ctx, credentials, projectName, triggerID)
		if triggerCleanupErr != nil {
			logger.Error("error removing trigger", zap.Error(triggerCleanupErr))
			return "", triggerCleanupErr
		}
	}

	imageTag := sourceAppTagGCRRef.Context().Tag(branch).String()

	mapVars := make(map[string]string)

	for _, tuple := range sourceApp.Vars {
		mapVars[tuple.Key] = tuple.Value
	}

	for _, tuple := range sourceApp.ReviewAppsConfig.Vars {
		mapVars[tuple.Key] = tuple.Value
	}

	var vars []*model.Tuple

	for k, v := range mapVars {
		vars = append(vars, &model.Tuple{
			Key:   k,
			Value: v,
		})
	}

	racExclusions := sourceApp.ReviewAppsConfig.ExcludedResources
	var reviewAppExclusions []*model.Resource
	reviewAppExclusions = append(reviewAppExclusions, sourceApp.ExcludedResources...)
	for _, r := range racExclusions {
		var found bool
		for _, e := range sourceApp.ExcludedResources {
			if strings.EqualFold(e.Kind, r.Kind) && strings.EqualFold(e.Name, r.Name) {
				found = true
				break
			}
		}
		if !found {
			reviewAppExclusions = append(reviewAppExclusions, r)
		}
	}

	reviewApp := &model.TuberApp{
		CloudSourceRepo:   sourceApp.CloudSourceRepo,
		ImageTag:          imageTag,
		Name:              reviewAppName,
		Paused:            false,
		ReviewApp:         true,
		SlackChannel:      sourceApp.SlackChannel,
		SourceAppName:     sourceApp.Name,
		State:             nil,
		TriggerID:         triggerID,
		Vars:              vars,
		ExcludedResources: reviewAppExclusions,
	}

	err = db.SaveApp(reviewApp)
	if err != nil {
		logger.Error("error saving review app", zap.Error(err))

		triggerCleanupErr := deleteReviewAppTrigger(ctx, credentials, projectName, triggerID)
		teardownErr := db.DeleteApp(reviewApp)

		if teardownErr != nil {
			logger.Error("error tearing down review app resources", zap.Error(teardownErr))
			return "", teardownErr
		}

		if triggerCleanupErr != nil {
			logger.Error("error removing trigger", zap.Error(triggerCleanupErr))
			return "", triggerCleanupErr
		}

		return "", err
	}

	return reviewAppName, nil
}

func DeleteReviewApp(ctx context.Context, db *core.DB, reviewAppName string, credentials []byte, projectName string) error {
	app, err := db.App(reviewAppName)
	if err != nil {
		return fmt.Errorf("review app not found")
	}

	if app.TriggerID != "" {
		err = deleteReviewAppTrigger(ctx, credentials, projectName, app.TriggerID)
		if err != nil {
			return err
		}
	}

	return core.DestroyTuberApp(db, app)
}

// yoinked from https://gitlab.com/gitlab-org/gitlab-runner/-/blob/0e2ae0001684f681ff901baa85e0d63ec7838568/executors/kubernetes/util.go#L23
const (
	DNS1123NameMaximumLength         = 52 // quick loophole to handle resource names. Originally 63
	DNS1123NotAllowedCharacters      = "[^-a-z0-9]"
	DNS1123NotAllowedStartCharacters = "^[^a-z0-9]+"
)

// yoinked from https://gitlab.com/gitlab-org/gitlab-runner/-/blob/0e2ae0001684f681ff901baa85e0d63ec7838568/executors/kubernetes/util.go#L268
func makeDNS1123Compatible(name string) string {
	name = strings.ToLower(name)

	nameNotAllowedChars := regexp.MustCompile(DNS1123NotAllowedCharacters)
	name = nameNotAllowedChars.ReplaceAllString(name, "")

	nameNotAllowedStartChars := regexp.MustCompile(DNS1123NotAllowedStartCharacters)
	name = nameNotAllowedStartChars.ReplaceAllString(name, "")

	if len(name) > DNS1123NameMaximumLength {
		name = name[0:DNS1123NameMaximumLength]
	}

	return name
}

func ReviewAppName(appName string, branch string) string {
	return makeDNS1123Compatible(fmt.Sprintf("%s-%s", appName, branch))
}

func copyNamespace(sourceApp string, reviewApp string) error {
	resource, err := k8s.Get("namespace", sourceApp, sourceApp, "-o", "json")
	if err != nil {
		return err
	}
	resource, err = duplicateResource(resource, sourceApp, reviewApp)
	if err != nil {
		return err
	}
	err = k8s.Apply(resource, reviewApp)
	if err != nil {
		return err
	}
	return nil
}

func copyResources(kind string, sourceApp string, reviewApp string, args ...string) error {
	data, err := duplicatedResources(kind, sourceApp, reviewApp, args...)
	if err != nil {
		return err
	}
	for _, resource := range data {
		applyErr := k8s.Apply(resource, reviewApp)
		if applyErr != nil {
			return applyErr
		}
	}
	return nil
}

func duplicatedResources(kind string, sourceApp string, reviewApp string, args ...string) ([][]byte, error) {
	list, err := k8s.ListKind(kind, sourceApp, args...)
	if err != nil {
		return nil, err
	}
	var resources [][]byte
	for _, resource := range list.Items {
		replicated, replicationErr := duplicateResource(resource, sourceApp, reviewApp)
		if replicationErr != nil {
			return nil, replicationErr
		}
		resources = append(resources, replicated)
	}
	return resources, nil
}

var nonGenericMetadata = []string{"annotations", "creationTimestamp", "namespace", "resourceVersion", "selfLink", "uid"}

func duplicateResource(resource []byte, sourceApp string, reviewApp string) ([]byte, error) {
	unmarshalled := make(map[string]interface{})
	err := json.Unmarshal(resource, &unmarshalled)
	if err != nil {
		return nil, err
	}
	metadata := unmarshalled["metadata"]
	stringKeyMetadata, ok := metadata.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resource metadata could not be coerced into map[string]interface{} for duplication")
	}
	for _, key := range nonGenericMetadata {
		delete(stringKeyMetadata, key)
	}

	stringName, ok := stringKeyMetadata["name"].(string)
	if !ok {
		return nil, fmt.Errorf("resource name could not be coerced into string for potential replacement")
	}
	if strings.Contains(stringName, sourceApp) {
		renamed := strings.ReplaceAll(stringName, sourceApp, reviewApp)
		stringKeyMetadata["name"] = renamed
	}

	unmarshalled["metadata"] = stringKeyMetadata

	genericized, err := json.Marshal(unmarshalled)
	if err != nil {
		return nil, err
	}
	return genericized, nil
}

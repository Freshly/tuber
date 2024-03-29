package reviewapps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/google/go-containerregistry/pkg/name"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"

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

func CreateReviewApp(ctx context.Context, db *core.DB, l *zap.Logger, sourceApp *model.TuberApp, branch string, credentials []byte, projectName string) (string, error) {
	reviewAppName := ReviewAppName(sourceApp.Name, branch)

	if db.AppExists(reviewAppName) {
		return "", fmt.Errorf("review app already exists")
	}

	logger := l.With(
		zap.String("appName", sourceApp.Name),
		zap.String("reviewAppName", reviewAppName),
		zap.String("branch", branch),
	)

	logger.Info("creating review app")

	if sourceApp.ReviewAppsConfig == nil || !sourceApp.ReviewAppsConfig.Enabled {
		return "", fmt.Errorf("source app is not enabled for review apps")
	}

	if sourceApp.CloudSourceRepo == "" {
		return "", fmt.Errorf("cloudSourceRepo is blank; it's required for review app trigger creation")
	}

	sourceAppTagGCRRef, err := name.ParseReference(sourceApp.ImageTag)
	if err != nil {
		return "", fmt.Errorf("source app image tag misconfigured: %v", err)
	}

	logger.Info("creating review app resources")

	err = NewReviewAppSetup(sourceApp.Name, reviewAppName)
	if err != nil {
		return "", err
	}

	logger.Info("creating and running review app trigger")
	cloudbuildService, err := cloudbuild.NewService(ctx, option.WithCredentialsJSON(credentials))
	if err != nil {
		return "", fmt.Errorf("cloudbuild service: %w", err)
	}
	service := cloudbuild.NewProjectsTriggersService(cloudbuildService)
	repoSource := cloudbuild.RepoSource{
		BranchName: branch,
		ProjectId:  projectName,
		RepoName:   sourceApp.CloudSourceRepo,
	}

	var triggerID string
	call := service.List(projectName)
	allTriggers, listErr := call.Do()
	if listErr != nil {
		logger.Error("triggers list failed, skipping exists check")
	} else {
		for _, trigger := range allTriggers.Triggers {
			if trigger.Name == reviewAppName {
				triggerID = trigger.Id
			}
		}
	}

	if triggerID == "" {
		triggerID, err = CreateTrigger(service, repoSource, projectName, reviewAppName)
		if err != nil {
			logger.Error("failed to create trigger", zap.Error(err))
			return "", err
		}
	}

	if triggerID == "" {
		logger.Error("triggerID blank after exists check and create block")
		return "", fmt.Errorf("triggerID blank after exists check and create block")
	}

	logger = logger.With(zap.String("triggerId", triggerID))
	err = RunTrigger(service, repoSource, triggerID, projectName)
	if err != nil {
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

// Kubernetes & DNS1123 Rules
// contain at most 63 characters
// contain only lowercase alphanumeric characters or '-'
// start with an alphanumeric character
// end with an alphanumeric character
const (
	dns1123LabelFmt          = "([a-z0-9](?:[-a-z0-9]*[a-z0-9])?)"
	dns1123NameMaximumLength = 30 // quick loophole to handle resource names. Originally 63
	capitalLetters           = `[A-Z]`
	symbolsExclHyphen        = `[^-a-z0-9]`
)

var (
	dns1123LabelRe = regexp.MustCompile("^" + dns1123LabelFmt + "$")
	capitalsRe     = regexp.MustCompile(capitalLetters)
	symbolsRe      = regexp.MustCompile(symbolsExclHyphen)
)

// makeDNS1123Compatible utilizes the various constraints and regex above to selectively modify
// incoming text, only modifying as needed where invalid.
// Order of operations:
// 1. Trim to a suitable length
// 2. Lowercase all alpha characters
// 3. Replace underscores with dashes
// 4. Remove all disallowed symbols
// 5. Trim leading & trailing hyphens
// 6. Return a default based on the current time.
func makeDNS1123Compatible(name string) string {
	n := []byte(name)

	if len(n) > dns1123NameMaximumLength {
		n = n[0:dns1123NameMaximumLength]
	}
	for !dns1123LabelRe.Match(n) {
		switch {
		case capitalsRe.Match(n):
			n = bytes.ToLower(n)
		case bytes.Contains(n, []byte("_")):
			n = bytes.ReplaceAll(n, []byte("_"), []byte("-"))
		case symbolsRe.Match(n):
			n = symbolsRe.ReplaceAll(n, []byte(""))
		case bytes.Compare(n, []byte("-")) == 1:
			n = bytes.Trim(n, "-")
		default:
			// Default case will return name similar to 1631143096-review-app which is 21 characters.
			return fmt.Sprintf("%d-review-app", time.Now().Unix())
		}
	}

	return string(n)
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

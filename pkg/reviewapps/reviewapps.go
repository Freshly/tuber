package reviewapps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/k8s"

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

func CreateReviewApp(ctx context.Context, l *zap.Logger, branch string, appName string, credentials []byte, projectName string) (string, error) {
	reviewAppName := ReviewAppName(appName, branch)

	list, err := core.TuberReviewApps()
	if err != nil {
		return "", err
	}
	_, err = list.FindApp(reviewAppName)
	if err == nil {
		return "", fmt.Errorf("review app already exists")
	}

	logger := l.With(
		zap.String("appName", appName),
		zap.String("reviewAppName", reviewAppName),
		zap.String("branch", branch),
	)

	logger.Info("creating review app")
	sourceApp, err := core.FindApp(appName)
	if err != nil {
		return "", fmt.Errorf("can't find source app. is %s managed by tuber", appName)
	}

	logger.Info("creating review app resources")

	err = NewReviewAppSetup(appName, reviewAppName)
	if err != nil {
		logger.Error("error creating review app resources; tearing down", zap.Error(err))

		teardownErr := core.DestroyTuberApp(reviewAppName)
		if teardownErr != nil {
			logger.Error("error tearing down review app resources", zap.Error(teardownErr))
			return "", teardownErr
		}

		return "", err
	}

	logger.Info("creating app entry for review app")

	err = core.AddReviewAppConfig(reviewAppName, sourceApp.Repo, branch)
	if err != nil {
		teardownErr := core.DestroyTuberApp(reviewAppName)
		if teardownErr != nil {
			logger.Error("error tearing down review app resources", zap.Error(teardownErr))
			return "", teardownErr
		}

		return "", err
	}

	logger.Info("creating and running review app trigger")

	err = CreateAndRunTrigger(ctx, logger, credentials, sourceApp.Repo, projectName, reviewAppName, branch)
	if err != nil {
		logger.Error("error creating trigger; no trigger resource created", zap.Error(err))

		triggerCleanupErr := deleteReviewAppTrigger(ctx, credentials, projectName, reviewAppName)
		teardownErr := core.DestroyTuberApp(reviewAppName)
		cleanupConfigErr := core.RemoveReviewAppConfig(reviewAppName)

		if teardownErr != nil {
			logger.Error("error tearing down review app resources", zap.Error(teardownErr))
			return "", teardownErr
		}

		if cleanupConfigErr != nil {
			logger.Error("error removing config entry for app", zap.Error(cleanupConfigErr))
			return "", cleanupConfigErr
		}

		if triggerCleanupErr != nil {
			logger.Error("error removing trigger", zap.Error(triggerCleanupErr))
			return "", triggerCleanupErr
		}

		return "", err
	}
	return reviewAppName, nil
}

func DeleteReviewApp(ctx context.Context, reviewAppName string, credentials []byte, projectName string) error {
	_, err := core.FindReviewApp(reviewAppName)
	if err != nil {
		return fmt.Errorf("review app not found")
	}

	err = core.DestroyTuberApp(reviewAppName)
	if err != nil {
		return err
	}

	err = core.RemoveReviewAppConfig(reviewAppName)
	if err != nil {
		return err
	}

	return deleteReviewAppTrigger(ctx, credentials, projectName, reviewAppName)
}

func ReviewAppName(appName string, branch string) string {
	return fmt.Sprintf("%s-%s", url.QueryEscape(appName), url.QueryEscape(branch))
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

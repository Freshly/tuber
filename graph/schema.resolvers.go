package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/freshly/tuber/graph/generated"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/builds"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/db"
	"github.com/freshly/tuber/pkg/events"
	"github.com/freshly/tuber/pkg/gcr"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/reviewapps"
	"go.uber.org/zap"
)

func (r *mutationResolver) CreateApp(ctx context.Context, input model.AppInput) (*model.TuberApp, error) {
	err := core.NewAppSetup(input.Name, *input.IsIstio)
	if err != nil {
		return nil, err
	}

	inputApp := model.TuberApp{
		Name:     input.Name,
		ImageTag: *input.ImageTag,
	}

	if err := r.Resolver.db.SaveApp(&inputApp); err != nil {
		return nil, err
	}

	return &model.TuberApp{}, nil
}

func (r *mutationResolver) UpdateApp(ctx context.Context, input model.AppInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if input.ImageTag != nil {
		app.ImageTag = *input.ImageTag
	}

	if input.Paused != nil {
		app.Paused = *input.Paused
	}

	if err := r.Resolver.db.SaveApp(app); err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) RemoveApp(ctx context.Context, input model.AppInput) (*model.TuberApp, error) {
	app := &model.TuberApp{Name: input.Name}
	err := r.Resolver.db.DeleteApp(app)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (r *mutationResolver) Deploy(ctx context.Context, input model.AppInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	tag := app.ImageTag
	if input.ImageTag != nil {
		tag = *input.ImageTag
	}

	digest, err := gcr.DigestFromTag(tag, r.credentials)
	if err != nil {
		return nil, fmt.Errorf("unexpected error: couldn't find image for the tag: %v", err)
	}

	event := events.NewEvent(r.logger, digest, tag)
	if err != nil {
		return nil, fmt.Errorf("unexpected error: couldn't find image for the tag: %v", err)
	}

	go r.Resolver.processor.ReleaseApp(event, app)

	return app, nil
}

func (r *mutationResolver) DestroyApp(ctx context.Context, input model.AppInput) (*model.TuberApp, error) {
	err := reviewapps.DeleteReviewApp(ctx, r.Resolver.db, input.Name, r.Resolver.credentials, r.Resolver.projectName)

	if err != nil {
		return nil, err
	}

	return &model.TuberApp{Name: input.Name}, nil
}

func (r *mutationResolver) CreateReviewApp(ctx context.Context, input model.CreateReviewAppInput) (*model.TuberApp, error) {
	name, err := reviewapps.CreateReviewApp(ctx, r.Resolver.db, r.Resolver.logger, input.BranchName, input.Name, r.Resolver.credentials, r.Resolver.projectName)
	if err != nil {
		return nil, err
	}

	return &model.TuberApp{
		Name: name,
	}, nil
}

func (r *mutationResolver) SetAppVar(ctx context.Context, input model.SetTupleInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if app.Vars == nil {
		app.Vars = make([]*model.Tuple, 0)
	}

	found := false

	for _, tuple := range app.Vars {
		if tuple.Key == input.Key {
			tuple.Value = input.Value
			found = true
		}
	}

	if !found {
		app.Vars = append(app.Vars, &model.Tuple{Key: input.Key, Value: input.Value})
	}

	if err := r.Resolver.db.SaveApp(app); err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) UnsetAppVar(ctx context.Context, input model.SetTupleInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if app.Vars == nil {
		return app, nil
	}

	var vars []*model.Tuple
	for _, tuple := range app.Vars {
		if tuple.Key != input.Key {
			vars = append(vars, tuple)
		}
	}
	app.Vars = vars

	if err := r.Resolver.db.SaveApp(app); err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) SetAppEnv(ctx context.Context, input model.SetTupleInput) (*model.TuberApp, error) {
	mapName := fmt.Sprintf("%s-env", input.Name)

	if err := k8s.PatchSecret(mapName, input.Name, input.Key, input.Value); err != nil {
		return nil, err
	}

	if err := k8s.Restart("deployments", input.Name); err != nil {
		return nil, err
	}

	return &model.TuberApp{Name: input.Name}, nil
}

func (r *mutationResolver) UnsetAppEnv(ctx context.Context, input model.SetTupleInput) (*model.TuberApp, error) {
	mapName := fmt.Sprintf("%s-env", input.Name)

	if err := k8s.RemoveSecretEntry(mapName, input.Name, input.Key); err != nil {
		return nil, err
	}

	if err := k8s.Restart("deployments", input.Name); err != nil {
		return nil, err
	}

	return &model.TuberApp{Name: input.Name}, nil
}

func (r *mutationResolver) SetExcludedResource(ctx context.Context, input model.SetResourceInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.AppName)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	res := &model.Resource{Name: input.Name, Kind: input.Kind}
	app.ExcludedResources = append(app.ExcludedResources, res)

	if err := r.Resolver.db.SaveApp(app); err != nil {
		return nil, err
	}

	return app, nil
}

func (r *mutationResolver) UnsetExcludedResource(ctx context.Context, input model.SetResourceInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.AppName)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if app.ExcludedResources == nil {
		return app, nil
	}

	resources := []*model.Resource{}
	for _, rs := range app.ExcludedResources {
		if !(rs.Name == input.Name && rs.Kind == input.Kind) {
			resources = append(resources, rs)
		}
	}
	app.ExcludedResources = resources

	if err := r.Resolver.db.SaveApp(app); err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) Rollback(ctx context.Context, input model.AppInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if app.State == nil || len(app.State.Previous) == 0 {
		return nil, fmt.Errorf("no previous successful release found")
	}

	type decodedResource struct {
		decoded  []byte
		resource *model.Resource
	}

	type rollbackErr struct {
		err      error
		resource *model.Resource
	}

	var errors []rollbackErr
	var decodedResources []decodedResource
	for _, resource := range app.State.Previous {
		decoded, decodeErr := base64.StdEncoding.DecodeString(resource.Encoded)
		if decodeErr != nil {
			errors = append(errors, rollbackErr{err: decodeErr, resource: resource})
			continue
		}
		decodedResources = append(decodedResources, decodedResource{decoded: decoded, resource: resource})
	}

	if len(errors) != 0 {
		combined := "no rollback performed, errors decoding resources: "
		for _, decodeErr := range errors {
			combined = combined + fmt.Sprintf("%s:%s, ", decodeErr.resource.Kind, decodeErr.resource.Name)
		}
		return nil, fmt.Errorf(strings.TrimSuffix(combined, ", "))
	}

	for _, resource := range decodedResources {
		applyErr := k8s.Apply(resource.decoded, app.Name)
		if applyErr != nil {
			r.logger.Debug("rollback apply error", zap.Error(applyErr))
			errors = append(errors, rollbackErr{err: applyErr, resource: resource.resource})
			continue
		}
	}

	if len(errors) != 0 {
		combined := "partial rollback performed, errors applying resources: "
		for _, applyErr := range errors {
			combined = combined + fmt.Sprintf("%s:%s, ", applyErr.resource.Kind, applyErr.resource.Name)
		}
		return nil, fmt.Errorf(strings.TrimSuffix(combined, ", "))
	}

	return app, nil
}

func (r *mutationResolver) SetGithubRepo(ctx context.Context, input model.AppInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if input.GithubRepo == nil {
		return nil, fmt.Errorf("GithubRepo required for SetGithubRepo")
	}

	app.GithubRepo = *input.GithubRepo

	err = r.Resolver.db.SaveApp(app)
	if err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) SetCloudSourceRepo(ctx context.Context, input model.AppInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if input.CloudSourceRepo == nil {
		return nil, fmt.Errorf("CloudSourceRepo required for SetCloudSourceRepo")
	}

	app.CloudSourceRepo = *input.CloudSourceRepo

	err = r.Resolver.db.SaveApp(app)
	if err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) SetSlackChannel(ctx context.Context, input model.AppInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if input.SlackChannel == nil {
		return nil, fmt.Errorf("SlackChannel required for SetSlackChannel")
	}

	app.SlackChannel = *input.SlackChannel

	err = r.Resolver.db.SaveApp(app)
	if err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) ManualApply(ctx context.Context, input model.ManualApplyInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("error finding app, nothing applied")
		}

		return nil, fmt.Errorf("error finding app, nothing applied: %v", err)
	}

	var resources []string
	for _, resource := range input.Resources {
		if resource == nil {
			return nil, errors.New("nil string pointer found in resources, nothing applied")
		}

		decoded, decodeErr := base64.StdEncoding.DecodeString(*resource)
		if decodeErr != nil {
			r.logger.Error(decodeErr.Error())
			return nil, fmt.Errorf("decode error, nothing applied: %v", err)
		}
		resources = append(resources, string(decoded))
	}

	branch, err := gcr.TagFromRef(app.ImageTag)
	if err != nil {
		r.logger.Error(err.Error())
		return nil, fmt.Errorf("unable to parse app's current image tag, nothing applied: %v", err)
	}

	var gitSha string
	for _, tag := range app.CurrentTags {
		if branch != tag {
			gitSha = tag
			break
		}
	}

	if gitSha == "" {
		return nil, fmt.Errorf("git sha not found in current tags, nothing applied: %v", err)
	}

	digest, err := gcr.DigestFromTag(gitSha, r.Resolver.credentials)
	if err != nil {
		r.logger.Error(err.Error())
		return nil, fmt.Errorf("unable to pull digest from git sha, nothing applied: %v", err)
	}

	imageTagWithDigest, err := gcr.SwapTags(app.ImageTag, digest)
	if err != nil {
		r.logger.Error(err.Error())
		return nil, fmt.Errorf("unable to parse currently deployed image, nothing applied: %v", err)
	}

	err = core.BypassReleaser(app, imageTagWithDigest, resources, r.Resolver.processor.ClusterData)
	if err != nil {
		r.logger.Error(err.Error())
		return nil, err
	}

	return app, nil
}

func (r *mutationResolver) SetRacEnabled(ctx context.Context, input model.SetRacEnabledInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	app.ReviewAppsConfig.Enabled = input.Enabled

	err = r.Resolver.db.SaveApp(app)
	if err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) SetRacVar(ctx context.Context, input model.SetTupleInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if input.Key == "" {
		return nil, fmt.Errorf("key required for SetRacVar")
	}

	if input.Value == "" {
		return nil, fmt.Errorf("value required for SetRacVar")
	}

	rac := app.ReviewAppsConfig
	vars := rac.Vars

	var found bool
	for _, t := range vars {
		if t.Key == input.Key {
			t.Value = input.Value
			found = true
		}
	}
	if !found {
		vars = append(vars, &model.Tuple{Key: input.Key, Value: input.Value})
	}
	rac.Vars = vars
	app.ReviewAppsConfig = rac

	err = r.Resolver.db.SaveApp(app)
	if err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) UnsetRacVar(ctx context.Context, input model.SetTupleInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if input.Key == "" {
		return nil, fmt.Errorf("key required for UnsetRacVar")
	}

	rac := app.ReviewAppsConfig
	vars := rac.Vars

	var updatedVars []*model.Tuple
	for _, t := range vars {
		if t.Key != input.Key {
			updatedVars = append(vars, t)
		}
	}
	rac.Vars = updatedVars
	app.ReviewAppsConfig = rac

	err = r.Resolver.db.SaveApp(app)
	if err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) SetRacExclusion(ctx context.Context, input model.SetResourceInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.AppName)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if input.Name == "" {
		return nil, fmt.Errorf("resource name required for SetRacExclusion")
	}

	if input.Kind == "" {
		return nil, fmt.Errorf("resource kind required for SetRacExclusion")
	}

	rac := app.ReviewAppsConfig
	exclusions := rac.ExcludedResources

	for _, t := range exclusions {
		if strings.EqualFold(t.Name, input.Name) && strings.EqualFold(t.Kind, input.Kind) {
			return app, nil
		}
	}

	exclusions = append(exclusions, &model.Resource{Name: input.Name, Kind: input.Kind})
	rac.ExcludedResources = exclusions
	app.ReviewAppsConfig = rac

	err = r.Resolver.db.SaveApp(app)
	if err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *mutationResolver) UnsetRacExclusion(ctx context.Context, input model.SetResourceInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.AppName)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	if input.Name == "" {
		return nil, fmt.Errorf("resource name required for UnsetRacExclusion")
	}

	if input.Kind == "" {
		return nil, fmt.Errorf("resource kind required for UnsetRacExclusion")
	}

	rac := app.ReviewAppsConfig
	exclusions := rac.ExcludedResources

	var updatedExclusions []*model.Resource
	for _, t := range exclusions {
		if !strings.EqualFold(t.Name, input.Name) && !strings.EqualFold(t.Kind, input.Kind) {
			updatedExclusions = append(updatedExclusions, t)
		}
	}

	rac.ExcludedResources = updatedExclusions
	app.ReviewAppsConfig = rac

	err = r.Resolver.db.SaveApp(app)
	if err != nil {
		return nil, fmt.Errorf("could not save changes: %v", err)
	}

	return app, nil
}

func (r *queryResolver) GetApp(ctx context.Context, name string) (*model.TuberApp, error) {
	return r.Resolver.db.App(name)
}

func (r *queryResolver) GetApps(ctx context.Context) ([]*model.TuberApp, error) {
	return r.Resolver.db.SourceApps()
}

func (r *queryResolver) GetClusterInfo(ctx context.Context) (*model.ClusterInfo, error) {
	return &model.ClusterInfo{
		Name:              r.Resolver.clusterName,
		Region:            r.Resolver.clusterRegion,
		ReviewAppsEnabled: r.Resolver.reviewAppsEnabled,
	}, nil
}

func (r *tuberAppResolver) ReviewApps(ctx context.Context, obj *model.TuberApp) ([]*model.TuberApp, error) {
	return r.db.ReviewAppsFor(obj)
}

func (r *tuberAppResolver) Env(ctx context.Context, obj *model.TuberApp) ([]*model.Tuple, error) {
	mapName := fmt.Sprintf("%s-env", obj.Name)
	config, err := k8s.GetConfigResource(mapName, obj.Name, "Secret")

	if err != nil {
		return nil, err
	}

	list := make([]*model.Tuple, 0)

	for k, ev := range config.Data {
		v, err := base64.StdEncoding.DecodeString(ev)

		if err != nil {
			return nil, fmt.Errorf("could not decode value for %s: %v", k, err)
		}

		list = append(list, &model.Tuple{Key: k, Value: string(v)})
	}

	return list, nil
}

func (r *tuberAppResolver) CloudBuildStatuses(ctx context.Context, obj *model.TuberApp) ([]*model.Build, error) {
	builds, err := builds.FindByApp(obj, "freshly-docker")
	if err != nil {
		return nil, err
	}

	return builds, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// TuberApp returns generated.TuberAppResolver implementation.
func (r *Resolver) TuberApp() generated.TuberAppResolver { return &tuberAppResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type tuberAppResolver struct{ *Resolver }

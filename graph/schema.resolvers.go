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

func (r *mutationResolver) Deploy(ctx context.Context, input model.DeployInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("could not find app")
		}

		return nil, fmt.Errorf("unexpected error while trying to find app: %v", err)
	}

	tag := app.ImageTag
	if input.Tag != nil {
		tag = *input.Tag
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

func (r *mutationResolver) ExcludedResources(ctx context.Context) ([]*model.Resource, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) Rollback(ctx context.Context, input model.AppNameInput) (*model.TuberApp, error) {
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
		applyErr := k8s.Apply(resource.decoded, resource.resource.Name)
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

func (r *queryResolver) GetApp(ctx context.Context, name string) (*model.TuberApp, error) {
	return r.Resolver.db.App(name)
}

func (r *queryResolver) GetApps(ctx context.Context) ([]*model.TuberApp, error) {
	return r.Resolver.db.SourceApps()
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

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// TuberApp returns generated.TuberAppResolver implementation.
func (r *Resolver) TuberApp() generated.TuberAppResolver { return &tuberAppResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type tuberAppResolver struct{ *Resolver }

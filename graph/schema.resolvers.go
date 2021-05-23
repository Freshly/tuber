package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/freshly/tuber/graph/generated"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/db"
	"github.com/freshly/tuber/pkg/reviewapps"
)

func (r *mutationResolver) CreateApp(ctx context.Context, input *model.AppInput) (*model.TuberApp, error) {
	err := core.NewAppSetup(input.Name, input.IsIstio)
	if err != nil {
		return nil, err
	}

	inputApp := model.TuberApp{
		Name:     input.Name,
		ImageTag: input.ImageTag,
	}

	if err := r.Resolver.db.SaveApp(&inputApp); err != nil {
		return nil, err
	}

	return &model.TuberApp{}, nil
}

func (r *mutationResolver) UpdateApp(ctx context.Context, key string, input *model.AppInput) (*model.TuberApp, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RemoveApp(ctx context.Context, key string) (*model.TuberApp, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DestroyApp(ctx context.Context, key string) (*model.TuberApp, error) {
	panic(fmt.Errorf("not implemented"))
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

func (r *mutationResolver) SetAppVar(ctx context.Context, input model.SetAppVarInput) (*model.TuberApp, error) {
	app, err := r.Resolver.db.App(input.Name)
	if err != nil {
		if errors.As(err, &db.NotFoundError{}) {
			return nil, errors.New("Could not find app.")
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
		return nil, fmt.Errorf("Could not save changes: %v", err)
	}

	return app, nil
}

func (r *queryResolver) GetApp(ctx context.Context, name string) (*model.TuberApp, error) {
	return r.Resolver.db.App(name)
}

func (r *queryResolver) GetApps(ctx context.Context) ([]*model.TuberApp, error) {
	return r.Resolver.db.Apps()
}

func (r *tuberAppResolver) ReviewApps(ctx context.Context, obj *model.TuberApp) ([]*model.TuberApp, error) {
	return r.db.ReviewAppsFor(obj)
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

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *mutationResolver) DeleteApp(ctx context.Context, appID string) (*model.TuberApp, error) {
	panic(fmt.Errorf("not implemented"))
}

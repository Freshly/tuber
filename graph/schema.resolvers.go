package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph/generated"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
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

func (r *queryResolver) GetApp(ctx context.Context, name string) (*model.TuberApp, error) {
	return r.Resolver.db.App(name)
}

func (r *queryResolver) GetApps(ctx context.Context) ([]*model.TuberApp, error) {
	return r.Resolver.db.Apps()
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *mutationResolver) DeleteApp(ctx context.Context, appID string) (*model.TuberApp, error) {
	panic(fmt.Errorf("not implemented"))
}

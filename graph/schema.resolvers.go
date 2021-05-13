package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph/generated"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
)

func (r *mutationResolver) CreateApp(ctx context.Context, input *model.AppInput) (*model.TuberApp, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateApp(ctx context.Context, appID string, input *model.AppInput) (*model.TuberApp, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteApp(ctx context.Context, appID string) (*model.TuberApp, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetApp(ctx context.Context, name string) (*model.TuberApp, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetApps(ctx context.Context) ([]*model.TuberApp, error) {
	appList, err := core.TuberSourceApps()

	if err != nil {
		return nil, err
	}

	var list []*model.TuberApp

	for _, n := range appList {
		list = append(list, &model.TuberApp{
			Name:     n.Name,
			ImageTag: n.ImageTag,
		})
	}

	return list, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

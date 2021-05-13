package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph/generated"
	"github.com/freshly/tuber/graph/model"
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
	// if err != nil {
	// 	return err
	// }

	// sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })

	// if jsonOutput {
	// 	out, err := json.Marshal(apps)

	// 	if err != nil {
	// 		return err
	// 	}

	// 	os.Stdout.Write(out)

	// 	return nil
	// }

	// table := tablewriter.NewWriter(os.Stdout)
	// table.SetHeader([]string{"Name", "Image"})
	// table.SetBorder(false)

	// for _, app := range apps {
	// 	table.Append([]string{app.Name, app.ImageTag})
	// }

	// table.Render()
	return nil, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

package core

import (
	"reflect"
	"testing"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/report"
	"go.uber.org/zap"
)

func Test_releaser_exclude(t *testing.T) {
	type fields struct {
		logger           *zap.Logger
		errorScope       report.Scope
		app              *model.TuberApp
		digest           string
		data             *ClusterData
		releaseYamls     []string
		prereleaseYamls  []string
		postreleaseYamls []string
		db               *DB
	}
	type args struct {
		res []appResource
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []appResource
	}{
		{
			fields: fields{
				app: &model.TuberApp{
					ExcludedResources: []*model.Resource{
						{
							Kind: "Deployment",
							Name: "tuber",
						},
					},
				},
			},
			name: "excludes resources",
			args: args{
				res: []appResource{
					{
						kind: "Deployment",
						name: "tuber",
					},
				},
			},
			want: []appResource{},
		},
		{
			fields: fields{
				app: &model.TuberApp{
					ExcludedResources: []*model.Resource{
						{
							Kind: "Deployment",
							Name: "tuber",
						},
					},
				},
			},
			name: "excludes resources",
			args: args{
				res: []appResource{
					{
						kind: "Deployment",
						name: "tuber",
					},
					{
						kind: "Deployment",
						name: "sidekiq",
					},
				},
			},
			want: []appResource{
				{
					kind: "Deployment",
					name: "sidekiq",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := releaser{
				logger:           tt.fields.logger,
				errorScope:       tt.fields.errorScope,
				app:              tt.fields.app,
				digest:           tt.fields.digest,
				data:             tt.fields.data,
				releaseYamls:     tt.fields.releaseYamls,
				prereleaseYamls:  tt.fields.prereleaseYamls,
				postreleaseYamls: tt.fields.postreleaseYamls,
				db:               tt.fields.db,
			}
			if got := r.exclude(tt.args.res); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("releaser.exclude() = %v, want %v", got, tt.want)
			}
		})
	}
}

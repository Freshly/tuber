package cmd

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetRacVarCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "var [app name] [var key] [(optional if --unset) var value]",
	Short:         "override default vars for yaml interpolation that review apps from this source app will be created with",
	Args:          cobra.RangeArgs(2, 3),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetRacVar,
}

func runAppsSetRacVar(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]
	varKey := args[1]
	var varValue string
	var gql string
	if appsSetRacVarUnsetFlag {
		gql = `
			mutation($input: SetTupleInput!) {
				unsetRacVar(input: $input) {
					name
				}
			}
		`
		input := &model.SetTupleInput{
			Name:  appName,
			Key:   varKey,
			Value: varValue,
		}

		var respData struct {
			unsetRacVar *model.TuberApp
		}

		return graphql.Mutation(context.Background(), gql, nil, input, &respData)
	}

	if len(args) != 3 {
		return fmt.Errorf("var value is required unless --unset")
	}
	varValue = args[2]
	gql = `
			mutation($input: SetTupleInput!) {
				setRacVar(input: $input) {
					name
				}
			}
		`

	input := &model.SetTupleInput{
		Name:  appName,
		Key:   varKey,
		Value: varValue,
	}

	var respData struct {
		setRacVar *model.TuberApp
	}

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

var appsSetRacVarUnsetFlag bool

func init() {
	appsSetRacVarCmd.Flags().BoolVar(&appsSetRacVarUnsetFlag, "unset", false, "unset rather than default set")
	appsSetReviewappsConfigCmd.AddCommand(appsSetRacVarCmd)
}

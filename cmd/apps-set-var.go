package cmd

import (
	"context"
	"fmt"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var appsSetVarCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "var [app name] [var key] [(optional if --unset) var value]",
	Short:         "set or unset yaml interpolation a var",
	Args:          cobra.RangeArgs(2, 3),
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetVar,
}

func runAppsSetVar(cmd *cobra.Command, args []string) error {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	appName := args[0]
	varKey := args[1]
	var varValue string
	var gql string
	if appsSetVarUnsetFlag {
		gql = `
			mutation($input: SetTupleInput!) {
				unsetAppVar(input: $input) {
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
			unsetAppVar *model.TuberApp
		}

		return graphql.Mutation(context.Background(), gql, nil, input, &respData)
	}

	if len(args) != 3 {
		return fmt.Errorf("var value is required unless --unset")
	}
	varValue = args[2]
	gql = `
			mutation($input: SetTupleInput!) {
				setAppVar(input: $input) {
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
		setAppVar *model.TuberApp
	}
	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

var appsSetVarUnsetFlag bool

func init() {
	appsSetVarCmd.Flags().BoolVar(&appsSetVarUnsetFlag, "unset", false, "unset rather than default set")
	appsSetCmd.AddCommand(appsSetVarCmd)
}

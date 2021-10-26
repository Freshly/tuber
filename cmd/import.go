package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "import [path-to-json.json]",
	Short:        "import exported version of an app. will not edit - `apps remove` and recreate with this if that's your use case",
	Args:         cobra.ExactArgs(1),
	PreRunE:      promptCurrentContext,
	RunE:         runImportCmd,
}

func init() {
	importCmd.Flags().StringVar(&importSourceAppNameFlag, "source", "", "copy namespace, rolebindings, and secrets from a source app")
	rootCmd.AddCommand(importCmd)
}

var importSourceAppNameFlag string

func runImportCmd(cmd *cobra.Command, args []string) error {
	contents, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	var app model.TuberApp
	err = json.Unmarshal(contents, &app)
	if err != nil {
		return err
	}

	remarshalled, err := json.Marshal(app)
	if err != nil {
		return err
	}

	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	input := &model.ImportAppInput{
		App:           string(remarshalled),
		SourceAppName: importSourceAppNameFlag,
	}

	var respData struct {
		importApp *model.TuberApp
	}

	gql := `
			mutation($input: ImportAppInput!) {
				importApp(input: $input) {
					name
				}
			}
		`

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

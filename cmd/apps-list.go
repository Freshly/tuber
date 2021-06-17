package cmd

import (
	"context"
	"encoding/json"
	"os"
	"sort"

	"github.com/freshly/tuber/graph/model"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var jsonOutput bool

var appsListCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "list",
	Short:         "List tuberapps",
	PreRunE:       displayCurrentContext,
	RunE:          runAppsListCmd,
}

func runAppsListCmd(*cobra.Command, []string) (err error) {
	graphql, err := gqlClient()
	if err != nil {
		return err
	}

	gql := `
			query {
				getApps {
					name
					imageTag
				}
			}
		`

	var respData struct {
		GetApps []*model.TuberApp
	}

	if err := graphql.Query(context.Background(), gql, &respData); err != nil {
		return err
	}

	apps := respData.GetApps

	sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })

	if jsonOutput {
		out, err := json.Marshal(apps)

		if err != nil {
			return err
		}

		os.Stdout.Write(out)

		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Image"})
	table.SetBorder(false)

	for _, app := range apps {
		table.Append([]string{app.Name, app.ImageTag})
	}

	table.Render()
	return nil
}

func init() {
	appsListCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as json")
	appsCmd.AddCommand(appsListCmd)
}

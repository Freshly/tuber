package cmd

import (
	"os"
	"tuber/pkg/core"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var appsListCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "list",
	Short:        "List tuberapps",
	RunE:         list,
}

func list(*cobra.Command, []string) (err error) {
	apps, err := core.TuberApps()

	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Image"})
	table.SetBorder(false)

	for _, app := range apps {
		table.Append([]string{app.Name, app.ImageTag})
	}

	table.Render()
	return
}

func init() {
	appsCmd.AddCommand(appsListCmd)
}

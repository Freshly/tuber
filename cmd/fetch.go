package cmd

import (
	"fmt"
	"tuber/pkg/containers"
	"tuber/pkg/core"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fetchCmd)
}

var fetchCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "fetch [appName]",
	Short:        "Fetch Tuber yaml files",
	RunE:         fetch,
	Args:         cobra.ExactArgs(1),
	PreRunE:      displayCurrentContext,
}

func fetch(cmd *cobra.Command, args []string) (err error) {
	creds, err := credentials()
	if err != nil {
		return
	}

	apps, err := core.TuberApps()

	if err != nil {
		return
	}

	app, err := apps.FindApp(args[0])

	if err != nil {
		return
	}

	prerelease, yamls, err := containers.GetTuberLayer(app.GetRepositoryLocation(), creds)
	yamls = append(yamls, prerelease...)

	if err == nil {
		for i, yaml := range yamls {
			if i > 0 {
				fmt.Println("--")
			}
			fmt.Printf("%s\n", yaml)
		}
	}
	return
}

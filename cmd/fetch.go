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
	Short:        "display all tuber yaml files",
	RunE:         fetch,
	Args:         cobra.ExactArgs(1),
	PreRunE:      displayCurrentContext,
}

func fetch(cmd *cobra.Command, args []string) (err error) {
	creds, err := credentials()
	if err != nil {
		return
	}

	apps, err := core.TuberSourceApps()

	if err != nil {
		return
	}

	app, err := apps.FindApp(args[0])

	if err != nil {
		return
	}

	location := app.GetRepositoryLocation()

	sha, err := containers.GetLatestSHA(location, creds)

	if err != nil {
		return err
	}

	prerelease, yamls, err := containers.GetTuberLayer(app.GetRepositoryLocation(), sha, creds)
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

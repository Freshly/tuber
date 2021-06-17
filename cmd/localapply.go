package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/freshly/tuber/graph/model"
	"github.com/spf13/cobra"
)

var localApplyCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "local-apply [app name]",
	Short:        "DESIGNED FOR EMERGENCIES. pulls .tuber dir from WORKING DIRECTORY and immediately interpolates and applies it. currently requires the app to only push branch and sha tags",
	Long: `To ensure this does not change the deployed image (as it is only designed to address configuration issues),
it will assume the first tag from the LAST successful release it finds that ISNT the tag tuber watches
is a valid indication of a unique indicator within your docker image tags.
It will pull the digest matching that tag, and run apply - with no monitoring, notifications, alerting, or automated rollbacks.
So yes, as said above, DESIGNED FOR EMERGENCIES.
That said, the current state of the app will be updated, so the manual rollback command can undo whatever you're doing.`,
	Args:    cobra.ExactArgs(1),
	RunE:    runLocalApplyCmd,
	PreRunE: warnNuclearLaunchConfirmation,
}

func warnNuclearLaunchConfirmation(cmd *cobra.Command, args []string) error {
	fmt.Println(color.RedString("This command can be destructive, and should ONLY BE USED IN PRODUCTION IN CASE OF EMERGENCY."))
	fmt.Println(color.YellowString("Please ensure you are in the correct directory for the app you intend to modify."))
	fmt.Println(color.YellowString("And that you have the latest code for the app on the deployed branch."))
	fmt.Println(color.HiMagentaString("----- And that you know what you're doing. -----"))
	return promptCurrentContext(cmd, args)
}

func runLocalApplyCmd(cmd *cobra.Command, args []string) error {
	appName := args[0]

	var files []string
	err := filepath.Walk(".tuber", func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error prior to apply: %v", err)
	}

	var yamls []string
	for _, file := range files {
		if file == ".tuber" {
			continue
		}
		raw, readErr := ioutil.ReadFile(file)
		if readErr != nil {
			return readErr
		}
		yamls = append(yamls, string(raw))
	}

	if len(yamls) == 0 {
		return fmt.Errorf("no .tuber yamls found, nothing applied")
	}

	var encoded []*string
	for _, yaml := range yamls {
		e := base64.StdEncoding.EncodeToString([]byte(yaml))
		encoded = append(encoded, &e)
	}

	graphql, err := gqlClient()
	if err != nil {
		return fmt.Errorf("error prior to apply: %v", err)
	}

	gql := `
		mutation($input: ManualApplyInput!) {
			manualApply(input: $input) {
				name
			}
		}
	`

	input := &model.ManualApplyInput{
		Resources: encoded,
		Name:      appName,
	}

	var respData struct {
		deploy *model.TuberApp
	}

	return graphql.Mutation(context.Background(), gql, nil, input, &respData)
}

func init() {
	rootCmd.AddCommand(localApplyCmd)
}

package cmd

import (
	"fmt"
	"strings"
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "exec -a [appName] -w [specific workload] -c [specific container] [commands]",
	Short:        "execs a command on an app",
	RunE:         exec,
}

var workload string
var container string

func exec(cmd *cobra.Command, args []string) error {
	var workloadName string
	if workload != "" {
		workloadName = workload
	} else {
		workloadName = appName
	}

	var containerName string
	if container != "" {
		containerName = container
	} else {
		containerName = workloadName
	}

	template := `{{range $k, $v := $.spec.selector.matchLabels}}{{$k}}={{$v}},{{end}}`
	l, err := k8s.Get("deployment", workloadName, appName, "-o", "go-template", "--template", template)
	if err != nil {
		return err
	}

	labels := strings.TrimSuffix(string(l), ",")

	jsonPath := fmt.Sprintf(`-o=jsonpath="%s"`, `{.items[0].metadata.name}`)
	name, err := k8s.GetCollection("pods", appName, "-l", labels, jsonPath)
	if err != nil {
		return err
	}

	execArgs := []string{"-c", containerName}
	execArgs = append(execArgs, args...)

	err = k8s.Exec(strings.Trim(string(name), "\""), appName, execArgs...)

	return err
}

func init() {
	execCmd.Flags().StringVarP(&workload, "workload", "w", "", "specify a deployment name if it does not match your app's name")
	execCmd.Flags().StringVarP(&container, "container", "c", "", "specify a container (selects by the deployment name by default)")
	execCmd.Flags().StringVarP(&appName, "app", "a", "", "app name (required)")
	execCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(execCmd)
}

package cmd

import (
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "exec -a [appName] -w [specific workload] -c [specific container] [commands]",
	Short:        "executes a command on an app",
	RunE:         exec,
	PreRunE:      promptCurrentContext,
}

var container string

func exec(cmd *cobra.Command, args []string) error {
	var containerName string
	if container != "" {
		containerName = container
	} else {
		containerName = fetchWorkload()
	}
	podName, err := fetchPodname()
	if err != nil {
		return err
	}

	execArgs := []string{"-c", containerName}
	execArgs = append(execArgs, args...)

	return k8s.Exec(podName, appName, execArgs...)
}

func init() {
	execCmd.Flags().StringVarP(&workload, "workload", "w", "", "specify a deployment name if it does not match your app's name")
	execCmd.Flags().StringVarP(&pod, "pod", "p", "", "specify a pod (selects one randomly from deployment otherwise)")
	execCmd.Flags().StringVarP(&container, "container", "c", "", "specify a container (selects by the deployment name by default)")
	execCmd.Flags().StringVarP(&appName, "app", "a", "", "app name (required)")
	execCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(execCmd)
}

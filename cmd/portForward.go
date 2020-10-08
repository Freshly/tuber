package cmd

import (
	"tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

// portForwardCmd represents the portForward command
var portForwardCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "port-forward -a appName [port(s)]",
	Args:         cobra.MinimumNArgs(1),
	Short:        "forward requests from your local machine to a running pod",
	Long: `Forward requests from a local address and port to a running pod.
You are able to specify multiple addresses and ports, but all combinations must be valid and running.

This command will always run against a single pod until either that pod terminates or this command is closed.

Specifying a workload:
The workload name will default to the app name if not supplied. If the workload name is not the same as the app name, that argument will be required to run the command successfully.

For example. When the desired workload is a deployment named 'user-service-sidekiq' within a tuber app named 'user-service':
tuber port-forward -a user-service -w user-service-sidekiq 9292

Specifying pods:
If no pod name is supplied a pod will be randomly selected for you. To target a specific pod that can be supplied as an argument to '-p' or '--pod'.
`,
	RunE:    portForward,
	PreRunE: promptCurrentContext,
}

func portForward(cmd *cobra.Command, args []string) error {
	podName, err := fetchPodname()
	if err != nil {
		return err
	}
	ports := args

	portForwardArgs := []string{"--address", address, "--pod-running-timeout", podRunningTimeout}
	err = k8s.PortForward(podName, appName, ports, portForwardArgs...)
	return err
}

func init() {
	portForwardCmd.Flags().StringVar(&address, "address", "localhost", "specify an address on your local machine to forward from. Can be a comma separated list. Only IP addresses and 'localhost' are valid.")
	portForwardCmd.Flags().StringVar(&podRunningTimeout, "pod-running-timeout", "1m0s", "The length of time (like 5s, 2m, or 3h, higher than zero) to wait until at least one pod is running.")
	portForwardCmd.Flags().StringVarP(&workload, "workload", "w", "", "specify a deployment name if it does not match your app's name")
	portForwardCmd.Flags().StringVarP(&pod, "pod", "p", "", "specify a pod (selects one randomly from deployment otherwise)")
	portForwardCmd.Flags().StringVarP(&appName, "app", "a", "", "app name (required)")
	portForwardCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(portForwardCmd)
}

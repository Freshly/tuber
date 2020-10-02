package cmd

import (
	"github.com/spf13/cobra"
)

// portForwardCmd represents the portForward command
var portForwardCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "port-forward",
	Short:        "forward requests from your local machine to a running pod",
	Long: `Forward requests from a local address and port to a running pod.
You are able to specify multiple addresses and ports, but all combinations must be valid and running.

This command will always run against a single pod until either that pod terminates or this command is closed.

Specifying workloads or containers:
Container names and workload names will default to the app name if not supplied. If either container name or workload name are not the same as the app name, that argument will be required to run the command successfully.

For example. When the desired workload is a deployment named 'user-service-sidekiq' within a tuber app named 'user-service':
tuber exec -a user-service -w user-service-sidekiq

Specifying pods:
If no pod name is supplied a pod will be randomly selected for you. To target a specific pod that can be supplied as an argument to '-p' or '--pod'.
`,
	RunE:    portForward,
	PreRunE: promptCurrentContext,
}

func portForward(cmd *cobra.Command, args []string) error {
	return nil
}

func init() {
	portForwardCmd.Flags().StringVarP(&workload, "workload", "w", "", "specify a deployment name if it does not match your app's name")
	portForwardCmd.Flags().StringVarP(&container, "container", "c", "", "specify a container (selects by the deployment name by default)")
	portForwardCmd.Flags().StringVarP(&appName, "app", "a", "", "app name (required)")
	portForwardCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(portForwardCmd)
}

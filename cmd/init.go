package cmd

import (
	"fmt"

	"github.com/freshly/tuber/pkg/core"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [appName] [routePrefix] [--istio=serviceType]",
	Short: "initialize a .tuber directory and relevant yamls",
	Long: `creates and interpolates tuber yamls. specify a service type as one of type
(grpc grpc http http2 https mongo mysql redis tcp tls udp)`,
	SilenceUsage: true,
	Args:         cobra.ExactArgs(2),
	RunE:         initialize,
}

var istioType string

var istioServiceTypes = map[string]bool{
	"grpc":     true,
	"grpc-web": true,
	"http":     true,
	"http2":    true,
	"https":    true,
	"mongo":    true,
	"mysql":    true,
	"redis":    true,
	"tcp":      true,
	"tls":      true,
	"udp":      true,
}

func initialize(cmd *cobra.Command, args []string) error {
	appName := args[0]
	routePrefix := args[1]

	if istioType == "false" {
		return core.InitTuberApp(appName, routePrefix, false, istioType)
	}

	if !istioServiceTypes[istioType] {
		return fmt.Errorf("unsupported istio service type, see https://istio.io/docs/ops/configuration/traffic-management/protocol-selection/ for available options")
	}

	return core.InitTuberApp(appName, routePrefix, true, istioType)
}

func init() {
	initCmd.Flags().StringVar(&istioType, "istio", "false", "disable istio with `false`, otherwise define service type")
	initCmd.MarkFlagRequired("istio")
	rootCmd.AddCommand(initCmd)
}

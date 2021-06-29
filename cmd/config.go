package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/freshly/tuber/pkg/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "config",
	Short:         "open local tuber config in your default editor",
	Args:          cobra.NoArgs,
	RunE:          openConfig,
}

var defaultTuberConfig = `# clusters:
#   - shorthand: some-shorthand-name
#     name: fully_qualified_gke_cluster_name <- run 'kubectl config current-context'
#     url: https://tuber-url-for-this-cluster-without-slash-tuber.com
#     auth:
#     	tuber_desktop_client_id: desktop client id specific to tuber
#     	tuber_desktop_client_secret: desktop client secret specific to tuber (not actually secret)
#     	iap_backend_web_client_id: client ID from the backend powering your cluster's IAP ingress
`

func openConfig(cmd *cobra.Command, args []string) error {
	configPath, notFoundErr := config.Path()
	if notFoundErr != nil {
		return notFoundErr
	}

	_, notFoundErr = os.Stat(configPath)
	if notFoundErr != nil {
		dir, err := config.Dir()
		if err != nil {
			return err
		}

		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}

		f, err := os.Create(configPath)
		if err != nil {
			return err
		}
		f.Write([]byte(defaultTuberConfig))
	}

	var command *exec.Cmd

	switch currentOS := runtime.GOOS; currentOS {
	case "darwin":
		command = exec.Command("open", configPath)
	case "linux":
		command = exec.Command("xdg-open", configPath)
	case "windows":
		psCommand := fmt.Sprintf("start %v", configPath)
		command = exec.Command("cmd", "/c", psCommand, "/w")
	default:
		return fmt.Errorf("unsupported os for auto-open, tuber config located at %v", configPath)
	}

	err := command.Run()
	if err != nil {
		vsCodeFallbackErr := exec.Command("code", configPath).Run()
		if vsCodeFallbackErr == nil {
			return nil
		}
		return fmt.Errorf("\nauto-open with `%s` and `code` failed; tuber config located at %v", command.Path, configPath)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)
}

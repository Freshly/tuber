package cmd

import (
	"fmt"
	"io"
	"net/http"
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
	RunE:          runConfigCmd,
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

func configGetFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	contents, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func runConfigCmd(cmd *cobra.Command, args []string) error {
	configPath, pathNotFoundErr := config.Path()
	if pathNotFoundErr != nil {
		return pathNotFoundErr
	}

	var err error
	var configFromUrl []byte
	if configFromUrlFlag != "" {
		configFromUrl, err = configGetFromUrl(configFromUrlFlag)
		if err != nil {
			return err
		}
	}

	conf, loadErr := config.Load()
	if loadErr != nil {
		var dir string
		dir, err = config.Dir()
		if err != nil {
			return err
		}

		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if conf != nil && configUpdateFlag {
		if conf.ConfigSourceUrl == "" {
			return fmt.Errorf("config source url not set")
		}
		configFromUrl, err = configGetFromUrl(conf.ConfigSourceUrl)
		if err != nil {
			return err
		}
	}

	if conf == nil || configUpdateFlag || configFromUrlFlag != "" {
		var data []byte

		if string(configFromUrl) == "" {
			data = []byte(defaultTuberConfig)
		} else {
			data = configFromUrl
		}

		err = os.WriteFile(configPath, data, os.ModePerm)
		if err != nil {
			return err
		}
		conf, err = config.Load()
		if err != nil {
			return err
		}
		return nil
	}

	var command *exec.Cmd

	fmt.Println("opening " + configPath + " in your default editor (or vscode if that doesn't work)...")
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

	err = command.Run()
	if err != nil {
		vsCodeFallbackErr := exec.Command("code", configPath).Run()
		if vsCodeFallbackErr == nil {
			return nil
		}
		return fmt.Errorf("\nauto-open with `%s` and `code` failed; tuber config located at %v", command.Path, configPath)
	}
	return nil
}

var configFromUrlFlag string
var configUpdateFlag bool

func init() {
	configCmd.Flags().StringVar(&configFromUrlFlag, "from-url", "", "pass in a url we'll curl for config contents")
	configCmd.Flags().BoolVar(&configUpdateFlag, "update", false, "re-pull config from the url you created it with, using --from-url")
	rootCmd.AddCommand(configCmd)
}

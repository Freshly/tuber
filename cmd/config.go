package cmd

import (
	"fmt"
	"os"
	osExec "os/exec"
	"runtime"

	"github.com/freshly/tuber/pkg/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "config",
	Short:        "open local tuber config in your default editor",
	Args:         cobra.NoArgs,
	RunE:         openConfig,
}

var defaultTuberConfig = `# clusters:
#   someShorthandName: 
#     name: some_full_cluster_name
#     url: that_clusters_tuber_url
`

func openConfig(cmd *cobra.Command, args []string) error {
	configPath, err := config.Path()
	if err != nil {
		return err
	}

	_, err = os.Stat(configPath)
	if err != nil {
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

	var command *osExec.Cmd

	switch currentOS := runtime.GOOS; currentOS {
	case "darwin":
		command = osExec.Command("open", configPath)
	case "linux":
		command = osExec.Command("xdg-open", configPath)
	case "windows":
		psCommand := fmt.Sprintf("start %v", configPath)
		command = osExec.Command("cmd", "/c", psCommand, "/w")
	default:
		return fmt.Errorf("unsupported os for auto-open, tuber config located at %v", configPath)
	}

	out, err := command.CombinedOutput()
	if err != nil {
		err = fmt.Errorf(string(out)+"\n"+err.Error()+"\ntuber config located at %v", configPath)
	}
	return err
}

func init() {
	rootCmd.AddCommand(configCmd)
}

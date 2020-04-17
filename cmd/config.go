package cmd

import (
	"fmt"
	"os"
	osExec "os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "config",
	Short:        "open local tuber config in your default editor",
	Args:         cobra.NoArgs,
	RunE:         config,
}

var defaultTuberConfig = `# clusters:
#   someShorthandName: some_full_cluster_name
`

func config(cmd *cobra.Command, args []string) error {
	configPath, err := tuberConfigPath()
	if err != nil {
		return err
	}

	_, err = os.Stat(configPath)
	if err != nil {
		dir, err := tuberConfigDir()
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
		panic(fmt.Errorf("what are you on, plan9"))
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

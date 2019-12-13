package apply

import (
	"io"
	"os/exec"
	"tuber/pkg/util"
)

func Apply(yamls []util.Yaml) (out []byte, err error) {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}

	go func() {
		defer stdin.Close()
		lastIndex := len(yamls) - 1
		for i, yaml := range yamls {
			io.WriteString(stdin, yaml.Content)
			if i < lastIndex {
				io.WriteString(stdin, "---\n")
			}
		}
	}()

	out, err = cmd.CombinedOutput()
	if err != nil {
		return
	}

	return
}

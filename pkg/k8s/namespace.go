package k8s

import (
	"fmt"
	"os/exec"
)

// CreateNamespace create a new namespace in kubernetes
func CreateNamespace(namespace string) (err error) {
	cmd := exec.Command("kubectl", "create", "namespace", namespace)

	out, err := cmd.CombinedOutput()

	if cmd.ProcessState.ExitCode() != 0 {
		err = fmt.Errorf(string(out))
	}

	return
}

// BindNamespace create a new namespace in kubernetes
func BindNamespace(namespace string) (err error) {
	cmd := exec.Command("kubectl", "create", "rolebinding", "default-edit", "--clusterrole=edit", "--serviceaccount=tuber:default", "--namespace", namespace)

	out, err := cmd.CombinedOutput()

	if cmd.ProcessState.ExitCode() != 0 {
		err = fmt.Errorf(string(out))
	}

	return
}
package k8s

import (
	"os/exec"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func runKubectl(cmd *exec.Cmd) (out []byte, err error) {
	out, err = cmd.CombinedOutput()

	if viper.GetBool("debug") {
		logger, zapErr := zap.NewDevelopment()
		if zapErr != nil {
			return nil, zapErr
		}
		logger.Debug(string(out))
	}

	if err != nil || cmd.ProcessState.ExitCode() != 0 {
		err = newK8sError(out, err)
	}
	return
}

func kubectl(args ...string) ([]byte, error) {
	return runKubectl(exec.Command("kubectl", args...))
}

func pipeToKubectl(data []byte, args ...string) (out []byte, err error) {
	cmd := exec.Command("kubectl", args...)
	stdin, err := cmd.StdinPipe()

	if err != nil {
		return
	}

	_, err = stdin.Write(data)
	if err != nil {
		return
	}

	err = stdin.Close()
	if err != nil {
		return
	}

	return runKubectl(cmd)
}

// Apply `kubectl apply` data to a given namespace. Specify output or any other flags as args.
// Uses a stdin pipe to include the content of the data slice
func Apply(data []byte, namespace string, args ...string) (err error) {
	apply := []string{"apply", "-n", namespace, "-f", "-"}
	_, err = pipeToKubectl(data, append(apply, args...)...)
	return
}

// Get `kubectl get` a resource. Specify output or any other flags as args
func Get(kind string, name string, namespace string, args ...string) ([]byte, error) {
	get := []string{"get", kind, name, "-n", namespace}
	return kubectl(append(get, args...)...)
}

// Delete `kubectl delete` a resource. Specify output or any other flags as args
func Delete(kind string, name string, namespace string, args ...string) (err error) {
	deleteArgs := []string{"delete", kind, name, "-n", namespace}
	_, err = kubectl(append(deleteArgs, args...)...)
	return
}

// Create `kubectl create` a resource.
// Some resources take multiple args (like secrets), so both the resource type and any flags are the variadic
func Create(namespace string, resourceAndArgs ...string) (err error) {
	create := []string{"create", "-n", namespace}
	_, err = kubectl(append(create, resourceAndArgs...)...)
	return
}

// Restart runs a rollout restart on a given resource type for a namespace
// For example, `Restart("deployments", "some-app")` will restart _all_ deployments in that namespace
func Restart(resource string, namespace string, args ...string) (err error) {
	restart := []string{"rollout", "restart", resource, "-n", namespace}
	_, err = kubectl(append(restart, args...)...)
	return
}

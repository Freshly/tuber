package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	kubectlRoot string = "kubectl"
)

type Kubectl struct {
	Runner Runner
}

func NewKubectl() *Kubectl {
	return &Kubectl{Runner: &Run{}}
}

func (k *Kubectl) Get(kind string, namespace string, args ...string) ([]byte, error) {
	getArgs := []string{"get", kind, "-n", namespace}
	getArgs = append(getArgs, args...)

	cmd := exec.Command(kubectlRoot, getArgs...)
	out, err := k.Runner.Do(cmd)

	return []byte(strings.TrimSpace(string(out))), err
}

// Runner contains behavior around running commands directly
type Runner interface {
	Do(*exec.Cmd) ([]byte, error)
}

// Run utilizes the os/exec package to run commands
type Run struct{}

// Do attempts to run the command passed in, returning the data received back
// to the caller
func (r *Run) Do(cmd *exec.Cmd) ([]byte, error) {
	if cmd == nil {
		return nil, fmt.Errorf("received nil command")
	}

	return do(cmd)
}

func do(cmd *exec.Cmd) ([]byte, error) {
	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		reason := trimNewLine(stderr.String())
		return nil, fmt.Errorf("command execution failed with %s: %s", err, reason)
	}

	return stdout.Bytes(), nil
}

func trimNewLine(str string) string {
	return strings.TrimRight(str, "\n")
}

func runKubectl(cmd *exec.Cmd) ([]byte, error) {
	if viper.GetBool("TUBER_DEBUG") {
		logger, zapErr := zap.NewDevelopment()
		if zapErr != nil {
			return nil, zapErr
		}
		logger.Debug(strings.Join(cmd.Args, " "))
	}

	out, err := cmd.CombinedOutput()

	if err != nil || cmd.ProcessState.ExitCode() != 0 {
		err = newK8sError(out, err)
		return nil, err
	}

	if viper.GetBool("TUBER_DEBUG") {
		logger, zapErr := zap.NewDevelopment()
		if zapErr != nil {
			return nil, zapErr
		}
		logger.Debug(string(out))
	}

	return out, nil
}

func kubectl(args ...string) ([]byte, error) {
	return runKubectl(exec.Command("kubectl", args...))
}

func kubectlIO(args ...string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if viper.GetBool("TUBER_DEBUG") {
		logger, zapErr := zap.NewDevelopment()
		if zapErr != nil {
			return zapErr
		}
		logger.Debug(strings.Join(cmd.Args, " "))
	}

	return cmd.Run()
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

// GetCollection gets for plural resource types break if given even an empty name
func GetCollection(kind string, namespace string, args ...string) ([]byte, error) {
	get := []string{"get", kind, "-n", namespace}
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

// RolloutStatus waits and watches a rollout's progress
func RolloutStatus(kind string, name string, namespace string, timeout time.Duration, args ...string) error {
	status := []string{"rollout", "status", kind, name, "-n", namespace, "--timeout", timeout.String()}
	_, err := kubectl(append(status, args...)...)
	return err
}

// RolloutUndo runs undo on a rollout
func RolloutUndo(kind string, name string, namespace string, args ...string) error {
	status := []string{"rollout", "undo", kind, name, "-n", namespace}
	_, err := kubectl(append(status, args...)...)
	return err
}

// Exists tells you if a given resource already exists. Errors if a get call fails for any reason other than Not Found
func Exists(kind string, name string, namespace string, args ...string) (bool, error) {
	get := []string{"get", kind, name, "-n", namespace}
	_, err := kubectl(append(get, args...)...)
	if err, ok := err.(NotFoundError); ok {
		if ok {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Exec interactive exec into a pod with a series of args to run
func Exec(name string, namespace string, args ...string) error {
	execArgs := []string{"-n", namespace, "exec", "-it", name}
	execArgs = append(execArgs, args...)
	return kubectlIO(execArgs...)
}

// PortForward forward local requests to a running pod
func PortForward(podName string, namespace string, ports []string, args ...string) error {
	portForwardArgs := []string{"port-forward", podName}
	portForwardArgs = append(append(append(portForwardArgs, args...), ports...), "-n", namespace)
	return kubectlIO(portForwardArgs...)
}

type unmarshalledList struct {
	Items []interface{} `json:"items"`
}

// List represents a nested list of Items
type List struct {
	Items [][]byte
}

// ListKind returns a List resource, with an Items slice of raw yamls
func ListKind(kind string, namespace string, args ...string) (List, error) {
	get := []string{"get", kind, "-n", namespace, "-o", "json"}
	out, err := kubectl(append(get, args...)...)
	if err != nil {
		return List{}, err
	}
	var unmarshalled unmarshalledList
	err = json.Unmarshal(out, &unmarshalled)
	if err != nil {
		return List{}, err
	}
	var l List
	for _, resource := range unmarshalled.Items {
		marshalled, marshalErr := json.Marshal(resource)
		if marshalErr != nil {
			return List{}, marshalErr
		}
		l.Items = append(l.Items, marshalled)
	}
	return l, nil
}

// UseCluster switch current configured kubectl cluster
func UseCluster(cluster string) error {
	_, err := kubectl([]string{"config", "use-context", cluster}...)
	return err
}

// CanI is for authorization checks
func CanI(namespace string, verb string, objectType string, args ...string) (bool, error) {
	canDeploy := []string{"auth", "can-i", verb, objectType, "-n", namespace}
	out, err := kubectl(append(canDeploy, args...)...)
	if err != nil {
		return false, err
	}

	return strings.Trim(string(out), "\r\n") == "yes", nil
}

func CanIAllNamespaces(verb string, objectType string, args ...string) (bool, error) {
	canDeploy := []string{"auth", "can-i", verb, objectType, "--all-namespaces"}
	out, err := kubectl(append(canDeploy, args...)...)
	if err != nil {
		return false, err
	}

	return strings.Trim(string(out), "\r\n") == "yes", nil
}

// CurrentCluster the current configured kubectl cluster
func CurrentCluster() (string, error) {
	out, err := kubectl([]string{"config", "current-context"}...)
	if err != nil {
		return "", err
	}
	return strings.Trim(string(out), "\r\n"), nil
}

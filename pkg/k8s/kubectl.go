package k8s

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func runKubectl(cmd *exec.Cmd) ([]byte, error) {
	if viper.GetBool("debug") {
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

	if viper.GetBool("debug") {
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
	cmd := exec.Command("kubectl", execArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// UseCluster switch current configured kubectl cluster
func UseCluster(cluster string) error {
	_, err := kubectl([]string{"config", "use-context", cluster}...)
	return err
}

// CurrentCluster the current configured kubectl cluster
func CurrentCluster() (string, error) {
	out, err := kubectl([]string{"config", "current-context"}...)
	if err != nil {
		return "", err
	}

	str := strings.Trim(string(out), "\r\n")
	return str, nil
}

// ClusterToken returns the auth key for the current cluster
func ClusterToken() (string, error) {
	type clusterAuth struct {
		AccessToken string `yaml:"access-token"`
	}

	type clusterAuthProvider struct {
		Config clusterAuth `yaml:"config"`
	}

	type userData struct {
		AuthProvider clusterAuthProvider `yaml:"auth-provider"`
	}

	type clusterUser struct {
		Name     string   `yaml:"name"`
		UserData userData `yaml:"user"`
	}

	type clusterUsers struct {
		Users []clusterUser `yaml:"users"`
	}

	var tmp clusterUsers

	out, err := kubectl([]string{"config", "view", "--raw"}...)
	if err != nil {
		return "", err
	}

	yaml.Unmarshal(out, &tmp)

	clusterName, err := CurrentCluster()
	if err != nil {
		return "", err
	}

	var authToken string
	for _, v := range tmp.Users {
		if v.Name == clusterName {
			authToken = v.UserData.AuthProvider.Config.AccessToken
			break
		}
	}

	if authToken == "" {
		return authToken, fmt.Errorf("no auth token found")
	}

	return authToken, nil
}

// CanDeploy determines if the current user can create a deployment
func CanDeploy(appName, token string) bool {
	t := fmt.Sprintf("--token=%s", token)

	out, err := kubectl([]string{"auth", "can-i", "create", "deployments", "-n", appName, t}...)
	if err != nil {
		return false
	}

	result := strings.Trim(string(out), "\r\n")

	return result == "yes"
}

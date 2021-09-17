package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"
	"go.uber.org/zap"
)

// RunPrerelease takes an array of pods, that are designed to be single use command runners
// that have access to the new code being released.
func RunPrerelease(logger *zap.Logger, resources []appResource, app *model.TuberApp) error {
	for _, resource := range resources {
		if resource.kind != "Pod" {
			return fmt.Errorf("prerelease resources must be Pods, received %s", resource.kind)
		}

		err := k8s.Apply(resource.contents, app.Name)
		if err != nil {
			return err
		}

		err = WaitForPhase(resource.name, "pod", app, resource.timeout)
		if err != nil {
			logger.Error("prerelease faled", zap.Error(err))
			contextErr := fmt.Errorf("prerelease phase failed for pod: %s", resource.name)
			deleteErr := k8s.Delete("pod", resource.name, app.Name)
			if deleteErr != nil {
				return fmt.Errorf(contextErr.Error() + "\n also failed delete:" + deleteErr.Error())
			}
			return contextErr
		}

		err = k8s.Delete("pod", resource.name, app.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func WaitForPhase(name string, kind string, app *model.TuberApp, resourceTimeout time.Duration) error {
	containerStatusesTemplate := fmt.Sprintf(
		`go-template="%s"`,
		"{{range .status.containerStatuses}}{{if .state.terminated.reason}}{{.state.terminated.reason}}{{end}}{{end}}",
	)

	phaseTemplate := fmt.Sprintf(`go-template="%s"`, "{{.status.phase}}")

	failureTemplate := fmt.Sprintf(
		`go-template="%s"`,
		"{{range .status.containerStatuses}}{{.state.terminated.message}}{{end}}",
	)
	timeout := time.Now().Add(time.Minute * 10)
	if resourceTimeout > 0 {
		timeout = time.Now().Add(resourceTimeout)
	}

	for {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout")
		}
		time.Sleep(10 * time.Second)

		statuses, err := k8s.Get(kind, name, app.Name, "-o", containerStatusesTemplate)
		if err != nil {
			return err
		}

		completedContainerFound := strings.Contains(strings.Trim(string(statuses), `"`), "Completed")
		if completedContainerFound {
			break
		}

		status, err := k8s.Get(kind, name, app.Name, "-o", phaseTemplate)
		if err != nil {
			return err
		}

		switch stringStatus := strings.Trim(string(status), `"`); stringStatus {
		case "Succeeded":
			return nil
		case "Failed":
			message, failedRetrieval := k8s.Get(kind, name, app.Name, "-o", failureTemplate)
			if err != nil {
				return failedRetrieval
			}
			return fmt.Errorf(string(message))
		default:
			continue
		}
	}

	return nil
}

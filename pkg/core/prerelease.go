package core

import (
	"fmt"
	"strings"
	"time"
	"tuber/pkg/k8s"
)

// RunPrerelease takes an array of pods, that are designed to be single use command runners
// that have access to the new code being released.
func RunPrerelease(resources []appResource, app *TuberApp) error {
	for _, resource := range resources {
		if resource.kind != "Pod" {
			return fmt.Errorf("prerelease resources must be Pods, received %s", resource.kind)
		}

		err := k8s.Apply(resource.contents, app.Name)
		if err != nil {
			return err
		}

		err = waitForPhase(resource.name, "pod", app, resource.timeout)
		if err != nil {
			deleteErr := k8s.Delete("pod", resource.name, app.Name)
			if deleteErr != nil {
				return fmt.Errorf(err.Error() + "\n also failed delete:" + deleteErr.Error())
			}
			return deleteErr
		}

		return k8s.Delete("pod", resource.name, app.Name)
	}
	return fmt.Errorf("unhandled prerelease run exit")
}

func waitForPhase(name string, kind string, app *TuberApp, resourceTimeout time.Duration) error {
	phaseTemplate := fmt.Sprintf(`go-template="%s"`, "{{.status.phase}}")
	failureTemplate := fmt.Sprintf(
		`go-template="%s"`,
		"{{range .status.containerStatuses}}{{.state.terminated.message}}{{end}}",
	)
	var timeout time.Time
	if resourceTimeout == 0 {
		timeout = time.Now().Add(time.Minute * 10)
	} else {
		timeout = time.Now().Add(resourceTimeout)
	}

	for {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout")
		}
		time.Sleep(5 * time.Second)

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
}

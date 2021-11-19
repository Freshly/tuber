package core

import (
	"encoding/json"
	"fmt"
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

		kc := k8s.NewKubectl()
		err = WaitForPhase(kc, resource.name, "pod", app, resource.timeout, 2*time.Second)
		if err != nil {
			logger.Error("prerelease failed", zap.Error(err))
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

type Terminated struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}
type PhaseData struct {
	Phase             string       `json:"phase"`
	ContainerStatuses []Terminated `json:"container-statuses,omitempty"`
}

const phaseTmpl string = `{"phase":"{{.status.phase}}"{{if .status.containerStatuses}},"container-statuses":[{{range $i, $status := .status.containerStatuses}}{{if $status.state.terminated.reason}}{{if $i}},{"reason":"{{$status.state.terminated.reason}}"{{if ne $status.state.terminated.reason "Completed"}},"message":"{{$status.state.terminated.message}}"{{end}}}{{else}}{"reason":"{{$status.state.terminated.reason}}"{{if and ($status.state.terminated.reason) (ne $status.state.terminated.reason "Completed")}},"message":"{{$status.state.terminated.message}}"{{end}}}{{end}}{{end}}{{end}}]{{end}}}`

// WaitForPhase calls to kubectl to retrieve the status of the prerelease pod and its container,
// waiting for it to Complete, Succeed, or Fail.
// TODO: Function doesn't need the entire app object so should be refactored to only take the name.
// The app name is what's used as the namespace value so the naming could also use being made a bit
// more communicative
// For ease of use and dependency injection, all functions in this file could probably use being
// turned into a method on a "prereleaser" struct. (being able to adjust the overall timeout would)
// be helpful.
func WaitForPhase(kbc *k8s.Kubectl, name string, kind string, app *model.TuberApp, resourceTimeout, checkDelay time.Duration) error {
	timeout := time.Now().Add(10 * time.Minute)
	if resourceTimeout > 0 {
		timeout = time.Now().Add(resourceTimeout)
	}

	for {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout")
		}
		time.Sleep(checkDelay)

		out, err := kbc.Get(kind, app.Name, name, "-o", "go-template="+phaseTmpl)
		if err != nil {
			return err
		}

		var phaseData PhaseData
		if err := json.Unmarshal(out, &phaseData); err != nil {
			return err
		}

		for _, cs := range phaseData.ContainerStatuses {
			if cs.Reason == "Completed" {
				break
			}
		}

		if phaseData.Phase == "Succeeded" {
			return nil
		}

		if phaseData.Phase == "Failed" {
			return fmt.Errorf("%v", phaseData.ContainerStatuses)
		}
	}
}

package k8s

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/argo-rollouts/pkg/kubectl-argo-rollouts/info"
)

func WatchArgoRollout(name string, namespace string, duration time.Duration) error {
	timeout := time.Now().Add(duration)
	for {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout waiting for healthy rollout")
		}
		time.Sleep(10 * time.Second)
		ready, err := argoRolloutStatus(name, namespace)
		if err != nil {
			return err
		}
		if ready {
			return nil
		}
	}
}

func argoRolloutStatus(name string, namespace string) (bool, error) {
	rollout, err := getRolloutResource(name, namespace)
	if err != nil {
		return false, err
	}

	status, message := info.RolloutStatusString(rollout)
	if status == "Healthy" {
		return true, nil
	} else if status == "Progressing" {
		return false, nil
	} else if status == "Paused" {
		return false, nil
	} else {
		return false, fmt.Errorf("unhealthy rollout status: %s, message: %s", status, message)
	}
}

// pulled from pkg/kubectl-argo-rollouts/cmd/abort/abort.go
// and pkg/kubectl-argo-rollouts/cmd/promote/promote.go
// no, this isn't good
const (
	abortPatch          = `{"status":{"abort":true}}`
	setCurrentStepPatch = `{
	"status": {
		"currentStepIndex": %d
	}
}`
)

func AbortArgoRollout(name string, namespace string) error {
	return Patch("rollout", name, namespace, abortPatch)
}

func ImmediateArgoRollout(name string, namespace string) error {
	rollout, err := getRolloutResource(name, namespace)
	if err != nil {
		return err
	}
	return Patch("rollout", name, namespace, fmt.Sprintf(setCurrentStepPatch, len(rollout.Spec.Strategy.Canary.Steps)))
}

func getRolloutResource(name string, namespace string) (*v1alpha1.Rollout, error) {
	out, err := Get("rollout", name, namespace, "-o", "json")
	if err != nil {
		return nil, err
	}

	var rollout v1alpha1.Rollout
	err = json.Unmarshal(out, &rollout)
	if err != nil {
		return nil, err
	}
	return &rollout, nil
}

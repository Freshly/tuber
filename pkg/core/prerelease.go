package core

import (
	"fmt"
	"strings"
	"time"
	"tuber/pkg/k8s"

	"github.com/goccy/go-yaml"
)

type prerelease struct {
	Kind     string
	Metadata Metadata
}

// Metadata exported for the yaml unmarshaller
type Metadata struct {
	Name string
}

// RunPrerelease takes an array of pods, that are designed to be single use command runners
// that have access to the new code being released.
func RunPrerelease(tubers []string, app *TuberApp, digest string, clusterData *ClusterData) error {
	for _, tuber := range tubers {
		interpolatedTuber, err := interpolate(tuber, tuberData(digest, app, clusterData))
		if err != nil {
			return err
		}
		prereleaser := prerelease{}
		err = yaml.Unmarshal([]byte(interpolatedTuber), &prereleaser)
		if err != nil {
			return err
		}

		if prereleaser.Kind != "Pod" {
			return fmt.Errorf("prerelease resources must be Pods, received %s", prereleaser.Kind)
		}

		err = k8s.Apply(interpolatedTuber, app.Name)
		if err != nil {
			return err
		}

		err = waitForPhase(prereleaser.Metadata.Name, "pod", app)
		if err != nil {
			deleteErr := k8s.Delete("pod", prereleaser.Metadata.Name, app.Name)
			if deleteErr != nil {
				return fmt.Errorf(err.Error() + "\n also failed delete:" + deleteErr.Error())
			}
			return deleteErr
		}

		return k8s.Delete("pod", prereleaser.Metadata.Name, app.Name)
	}
	return fmt.Errorf("unhandled prerelease run exit")
}

func waitForPhase(name string, kind string, app *TuberApp) error {
	phaseTemplate := fmt.Sprintf(`go-template="%s"`, "{{.status.phase}}")
	failureTemplate := fmt.Sprintf(
		`go-template="%s"`,
		"{{range .status.containerStatuses}}{{.state.terminated.message}}{{end}}",
	)
	timeout := time.Now().Add(time.Minute * 10)

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

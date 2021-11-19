package core

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"

	// imported as "a" since core package has a function called "assert" already.
	a "github.com/stretchr/testify/assert"
)

const (
	testResourceTimeout = time.Duration(1 * time.Millisecond)
	testCheckDelay      = time.Duration(1 * time.Nanosecond)
)

type wfpArgs struct {
	rName, rKind, aName string
}

func TestWaitForPhase(t *testing.T) {
	testCases := []struct {
		name         string
		args         wfpArgs
		delay        time.Duration
		expectedErr  error
		expectedK8s  string
		k8sResponses []string
		k8sErrs      error
	}{
		{
			name:        "wait for container to complete",
			args:        wfpArgs{rName: "test-resource", rKind: "pod", aName: "test-namespace"},
			expectedK8s: "kubectl get pod -n test-namespace test-resource -o go-template=" + phaseTmpl,
			k8sResponses: []string{
				`{"phase":"Running"}`,
				`{"phase":"Running"}`,
				`{"phase":"Succeeded","container-statuses":[{"reason":"Completed"}]}`,
			},
		},
		{
			name:         "timeout",
			args:         wfpArgs{rName: "test-resource", rKind: "pod", aName: "test-namespace"},
			delay:        2 * time.Millisecond,
			expectedK8s:  "kubectl get pod -n test-namespace test-resource -o go-template=" + phaseTmpl,
			expectedErr:  errors.New("timeout"),
			k8sResponses: []string{`{"phase":"Running"}`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ta := &model.TuberApp{Name: tc.args.aName}
			checkDelay := tc.delay
			if checkDelay.String() == "0s" {
				checkDelay = testCheckDelay
			}

			var count int
			kc := newTestKubectl(t, count, tc.expectedK8s, tc.k8sResponses, tc.k8sErrs)

			err := WaitForPhase(kc, tc.args.rName, tc.args.rKind, ta, testResourceTimeout, checkDelay)

			a.Equal(t, tc.expectedErr, err)
		})
	}
}

// newTestKubectl returns a pointer to a Kubectl with an overridden Runner to assist with
// testing commands being run without actually running them.
func newTestKubectl(t *testing.T, reqCount int, expectedCmd string, k8sResponses []string, k8sErrs error) *k8s.Kubectl {
	t.Helper()

	runnerFunc := func(c *exec.Cmd) ([]byte, error) {
		a.Equal(t, expectedCmd, stripCmdPath(c.String()))

		out := k8sResponses[reqCount]
		reqCount++

		return []byte(out), k8sErrs
	}

	return &k8s.Kubectl{
		Runner: runnerDoFunc(runnerFunc),
	}
}

type runnerDoFunc func(*exec.Cmd) ([]byte, error)

func (f runnerDoFunc) Do(cmd *exec.Cmd) ([]byte, error) {
	return f(cmd)
}

// stripCmdPath takes in a command's string value and strips any path that may be
// appended for simpler test assertions in case the command executable is at a different
// location
func stripCmdPath(cmd string) string {
	var out strings.Builder

	splitSpc := strings.Split(cmd, " ")
	splitSlash := strings.Split(splitSpc[0], "/")
	if len(splitSlash) == 1 {
		// early return if there's no path
		// prevents writing to builder
		return cmd
	}

	fmt.Fprintf(&out, splitSlash[len(splitSlash)-1])
	for _, s := range splitSpc[1:] {
		fmt.Fprintf(&out, " %s", s)
	}

	return out.String()
}

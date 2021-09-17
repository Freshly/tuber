package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/spf13/cobra"
)

var appsSetupCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "one-off [command] -a [appName]",
	Short:         "run a command on a temporary pod and watch for 5 minutes for the container to enter the completed status",
	PreRunE:       promptCurrentContext,
	RunE:          runAppsSetup,
}

var appsSetupPod = `apiVersion: v1
kind: Pod
metadata:
  name: {{.PodName}}
  annotations:
    sidecar.istio.io/inject: "false"
spec:
  restartPolicy: Never
  containers:
  - name: setup
    image: "{{.TuberImage}}"
    command: ["/bin/sh"]
    args: ["-c", "{{.Command}} > /dev/termination-log"]
    terminationMessagePolicy: FallbackToLogsOnError
    envFrom:
      - secretRef:
          name: "{{.TuberAppName}}-env"
`

type appsSetupTemplate struct {
	TuberAppName string
	TuberImage   string
	Command      string
	PodName      string
}

func runAppsSetup(cmd *cobra.Command, args []string) error {
	logger, err := createLogger()
	if err != nil {
		return err
	}
	defer logger.Sync()

	app, err := getApp(appNameFlag)
	if err != nil {
		return err
	}

	tpl, err := template.New("").Parse(appsSetupPod)
	if err != nil {
		return err
	}

	const podName = "setup"

	var buf bytes.Buffer
	err = tpl.Execute(&buf, appsSetupTemplate{TuberAppName: app.Name, TuberImage: app.ImageTag, Command: strings.Join(args, " "), PodName: podName})
	if err != nil {
		return err
	}
	err = k8s.Apply(buf.Bytes(), app.Name)
	if err != nil {
		return err
	}

	err = core.WaitForPhase(podName, "pod", app, 5*time.Minute)
	if err != nil {
		return err
	}

	if err != nil {
		deleteErr := k8s.Delete("pod", podName, app.Name)
		if deleteErr != nil {
			return fmt.Errorf(err.Error() + "\n also failed delete:" + deleteErr.Error())
		}
		return err
	}

	err = k8s.Delete("pod", podName, app.Name)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	appsSetupCmd.Flags().StringVarP(&appNameFlag, "app", "a", "", "app name")
	appsSetupCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(appsSetupCmd)
}

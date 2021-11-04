package cmd

import (
	"fmt"

	"github.com/freshly/tuber/pkg/k8s"

	"github.com/spf13/cobra"
)

var plantCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "plant [service account credentials path]",
	Short:        "install tuber to a cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         plant,
}

func plant(cmd *cobra.Command, args []string) error {
	existsAlready, err := k8s.Exists("namespace", "tuber", "tuber")
	if err != nil {
		return err
	}

	if existsAlready {
		return fmt.Errorf("tuber already planted")
	}

	var clusterRole = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: tuber-admin
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - '*'
`

	var clusterRoleBinding = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tuber-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tuber-admin
subjects:
- kind: ServiceAccount
  name: tuber
  namespace: tuber
`

	var serviceAccount = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: tuber
  namespace: tuber
`

	var namespace = `apiVersion: v1
kind: Namespace
metadata:
  name: tuber
  labels:
    istio-injection: false
`

	err = k8s.Apply([]byte(namespace), "tuber")
	if err != nil {
		return err
	}

	err = k8s.Apply([]byte(serviceAccount), "tuber")
	if err != nil {
		return err
	}

	err = k8s.Apply([]byte(clusterRole), "tuber")
	if err != nil {
		return err
	}

	err = k8s.Apply([]byte(clusterRoleBinding), "tuber")
	if err != nil {
		return err
	}

	err = k8s.CreateEnv("tuber")

	credentialsPath := args[0]
	err = k8s.CreateTuberCredentials(credentialsPath, "tuber")
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(plantCmd)
}

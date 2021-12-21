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
	PreRunE:      promptCurrentContext,
	RunE:         plant,
}

func setupAndAuth() error {
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
    istio-injection: "disabled"
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
	if err != nil {
		return err
	}

	err = k8s.CreateTuberCredentials("credentials.json", "tuber")
	if err != nil {
		return err
	}

	return nil
}

func firstDeploy(host string, gateway string) error {
	var everything = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: tuber
  namespace: tuber
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tuber
  namespace: tuber
  annotations:
    "tuber/rolloutTimeout": 30m
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tuber
  strategy:
    type: Recreate
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
      labels:
        app: tuber
    spec:
      serviceAccountName: tuber
      terminationGracePeriodSeconds: 1200
      containers:
        - image: "gcr.io/freshly-docker/tuber:main"
          name: tuber
          command: [ "tuber", "start" ]
          resources:
            requests:
              memory: "1Gi"
              cpu: "1"
            limits:
              memory: "1Gi"
              cpu: "1"
          volumeMounts:
            - name: tuber-credentials
              readOnly: true
              mountPath: "/etc/tuber-credentials"
            - name: tuber-bolt
              mountPath: "/etc/tuber-bolt"
          envFrom:
            - secretRef:
                name: tuber-env
          ports:
            - containerPort: 3000
      volumes:
        - name: tuber-credentials
          secret:
            secretName: tuber-credentials.json
        - name: tuber-bolt
          persistentVolumeClaim:
            claimName: tuber
---
apiVersion: v1
kind: Service
metadata:
  name: tuber
  namespace: tuber
spec:
  ports:
  - port: 3000
    name: http
  selector:
    app: tuber
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: tuber
  namespace: tuber
spec:
  hosts:
    - "%v"
  gateways:
    - "%v"
  http:
    - match:
        - uri:
            prefix: /tuber
      route:
        - destination:
            host: tuber
            port:
              number: 3000
`
	interpolated := fmt.Sprintf(everything, host, gateway)
	err := k8s.Apply([]byte(interpolated), "tuber")
	if err != nil {
		return err
	}

	return nil
}

func plant(cmd *cobra.Command, args []string) error {
	d, err := cmd.Flags().GetBool("deploy")
	if err != nil {
		return err
	}
	if d {
		var host string
		host, err = cmd.Flags().GetString("host")
		if err != nil {
			return err
		}

		var gateway string
		gateway, err = cmd.Flags().GetString("gateway")
		if err != nil {
			return err
		}

		if host == "" || gateway == "" {
			return fmt.Errorf("both the --host and --gateway flags are required along with deploy")
		}
		return firstDeploy(host, gateway)
	} else {
		return setupAndAuth()
	}
}

func init() {
	plantCmd.Flags().Bool("deploy", false, "deploy tuber the first time (run without this flag first)")
	plantCmd.Flags().String("host", "", "for the virtualservice")
	plantCmd.Flags().String("gateway", "", "for the virtualservice")
	rootCmd.AddCommand(plantCmd)
}

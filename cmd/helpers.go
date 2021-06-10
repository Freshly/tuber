package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	tuberbolt "github.com/freshly/tuber/pkg/db"
	"github.com/freshly/tuber/pkg/iap"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/report"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var address string
var appName string
var pod string
var podRunningTimeout string
var workload string

func db() (*core.DB, error) {
	var path string
	if _, err := os.Stat("/etc/tuber-bolt"); os.IsNotExist(err) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		path = wd + "/localbolt"
	} else {
		path = "/etc/tuber-bolt/db"
	}

	database, err := tuberbolt.NewDefaultDB(path, model.TuberApp{}.DBRoot())
	if err != nil {
		return nil, err
	}

	return core.NewDB(database), nil
}

// func getApp(appName string) (*model.TuberApp, error) {
// 	graphql := graph.NewClient(mustGetTuberConfig().CurrentClusterConfig().URL)
// 	gql := `
// 		query {
// 			getApp {
// 				cloudSourceRepo
// 				imageTag
// 				name
// 				paused
// 				reviewApp
// 				reviewAppsConfig{
// 					enabled
// 					vars {
// 						key
// 						value
// 					}
// 					skips {
// 						kind
// 						name
// 					}
// 				}
// 				slackChanne
// 				sourceAppName
// 				state {
// 					current {
// 						kind
// 						name
// 						encoded
// 					}
// 					previous {
// 						kind
// 						name
// 						encoded
// 					}
// 				}
// 				triggerID
// 				vars {
// 					key
// 					value
// 				}
// 			}
// 		}
// 	`

// 	var respData struct {
// 		GetApp *model.TuberApp
// 	}

// 	err := graphql.Query(context.Background(), gql, &respData)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if respData.GetApp == nil {
// 		return nil, fmt.Errorf("error retrieving app")
// 	}

// 	return respData.GetApp, nil
// }

func clusterData() (*core.ClusterData, error) {
	defaultGateway := viper.GetString("cluster-default-gateway")
	defaultHost := viper.GetString("cluster-default-host")
	adminGateway := viper.GetString("cluster-admin-gateway")
	adminHost := viper.GetString("cluster-admin-host")

	if defaultGateway == "" || defaultHost == "" || adminGateway == "" || adminHost == "" {
		config, err := k8s.GetSecret("tuber", "tuber-env")
		if err != nil {
			return nil, err
		}
		if defaultGateway == "" {
			defaultGateway = config.Data["TUBER_CLUSTER_DEFAULT_GATEWAY"]
		}
		if defaultHost == "" {
			defaultHost = config.Data["TUBER_CLUSTER_DEFAULT_HOST"]
		}
		if adminGateway == "" {
			adminGateway = config.Data["TUBER_CLUSTER_ADMIN_GATEWAY"]
		}
		if adminHost == "" {
			adminHost = config.Data["TUBER_CLUSTER_ADMIN_HOST"]
		}
	}

	data := &core.ClusterData{
		DefaultGateway: defaultGateway,
		DefaultHost:    defaultHost,
		AdminGateway:   adminGateway,
		AdminHost:      adminHost,
	}

	return data, nil
}

func credentials() ([]byte, error) {
	viper.SetDefault("credentials-path", "/etc/tuber-credentials/credentials.json")
	credentialsPath := viper.GetString("credentials-path")
	creds, err := ioutil.ReadFile(credentialsPath)

	if err != nil {
		config, err := k8s.GetSecret("tuber", "tuber-credentials.json")
		if err != nil {
			return nil, fmt.Errorf("error while running k8s.GetSecret: %v", err)
		}
		return []byte(config.Data["credentials.json"]), nil
	}

	return creds, nil
}

func checkAuth() error {
	if !iap.RefreshTokenExists() {
		return errors.New("tuber is not authorized. Please run `tuber auth`")
	}

	return nil
}

func promptCurrentContext(cmd *cobra.Command, args []string) error {
	if err := checkAuth(); err != nil {
		return err
	}

	skipConfirmation, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return err
	}

	if skipConfirmation {
		return nil
	}

	cluster, err := k8s.CurrentCluster()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "About to run %s on %s", cmd.Name(), cluster)
	fmt.Fprintf(os.Stderr, "\nPress ctrl+C to cancel, enter to continue...")
	var input string
	_, err = fmt.Scanln(&input)

	if err != nil {
		if err.Error() == "unexpected newline" {
			return nil
		} else if err.Error() == "EOF" {
			return fmt.Errorf("cancelled")
		} else {
			return err
		}
	}
	return nil
}

func displayCurrentContext(cmd *cobra.Command, args []string) error {
	if err := checkAuth(); err != nil {
		return err
	}

	skipConfirmation, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return err
	}

	if skipConfirmation {
		return nil
	}
	cluster, err := k8s.CurrentCluster()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Running %s on %s\n", cmd.Name(), cluster)

	return nil
}

func initErrorReporters() {
	report.ErrorReporters = []report.ErrorReporter{
		report.Sentry{
			Enable: viper.GetBool("sentry-enabled"),
			Options: sentry.ClientOptions{
				Dsn:              viper.GetString("sentry-dsn"),
				Environment:      viper.GetString("sentry-environment"),
				AttachStacktrace: true,
			},
		},
	}
	report.InitErrorReporters()
}

func fetchWorkload() string {
	if workload != "" {
		return workload
	}
	return appName
}

func fetchPodname() (string, error) {
	if pod != "" {
		return pod, nil
	}
	template := `{{range $k, $v := $.spec.selector.matchLabels}}{{$k}}={{$v}},{{end}}`
	l, err := k8s.Get("deployment", fetchWorkload(), appName, "-o", "go-template", "--template", template)
	if err != nil {
		return "", err
	}

	labels := strings.TrimSuffix(string(l), ",")

	jsonPath := fmt.Sprintf(`-o=jsonpath="%s"`, `{.items[0].metadata.name}`)
	podNameByte, err := k8s.GetCollection("pods", appName, "-l", labels, jsonPath)
	if err != nil {
		return "", err
	}
	return strings.Trim(string(podNameByte), "\""), nil
}

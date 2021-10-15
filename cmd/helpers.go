package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/freshly/tuber/graph"
	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/config"
	"github.com/freshly/tuber/pkg/core"
	tuberbolt "github.com/freshly/tuber/pkg/db"
	"github.com/freshly/tuber/pkg/iap"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/freshly/tuber/pkg/report"

	"github.com/briandowns/spinner"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var address string
var appNameFlag string
var pod string
var podRunningTimeout string
var workload string

func openDB() (*core.DB, error) {
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

func gqlClient() (*graph.GraphqlClient, error) {
	c, err := config.Load()
	if err != nil {
		return nil, err
	}

	cluster, err := c.CurrentClusterConfig()
	if err != nil {
		return nil, err
	}

	return graph.NewClient(cluster.URL, cluster.Auth.Audience), nil
}

func getApp(appName string) (*model.TuberApp, error) {
	graphql, err := gqlClient()
	if err != nil {
		return nil, err
	}

	gql := `
		query {
			getApp(name: "%s") {
				cloudSourceRepo
				createdAt
				updatedAt
				imageTag
				name
				paused
				reviewApp
				currentTags
				githubRepo
				reviewAppsConfig{
					enabled
					vars {
						key
						value
					}
					excludedResources {
						kind
						name
					}
				}
				slackChannel
				sourceAppName
				state {
					Current {
						kind
						name
						encoded
					}
					Previous {
						kind
						name
						encoded
					}
				}
				excludedResources {
					kind
					name
				}
				triggerID
				vars {
					key
					value
				}
			}
		}
	`

	var respData struct {
		GetApp *model.TuberApp
	}

	err = graphql.Query(context.Background(), fmt.Sprintf(gql, appName), &respData)
	if err != nil {
		return nil, err
	}

	if respData.GetApp == nil {
		return nil, fmt.Errorf("error retrieving app")
	}

	return respData.GetApp, nil
}

func clusterData() (*core.ClusterData, error) {
	defaultGateway := viper.GetString("TUBER_CLUSTER_DEFAULT_GATEWAY")
	defaultHost := viper.GetString("TUBER_CLUSTER_DEFAULT_HOST")
	adminGateway := viper.GetString("TUBER_CLUSTER_ADMIN_GATEWAY")
	adminHost := viper.GetString("TUBER_CLUSTER_ADMIN_HOST")

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
	viper.SetDefault("TUBER_CREDENTIALS_PATH", "/etc/tuber-credentials/credentials.json")
	credentialsPath := viper.GetString("TUBER_CREDENTIALS_PATH")
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

func checkAuth(audience string) error {
	exists, err := iap.RefreshTokenExists(audience)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("tuber is not authorized. Please run `tuber auth`")
	}

	return nil
}

func promptCurrentContext(cmd *cobra.Command, args []string) error {
	skipConfirmation, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return err
	}

	if skipConfirmation {
		return nil
	}

	c, err := config.Load()
	if err != nil {
		return err
	}

	cluster, err := c.CurrentClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to get current cluster context. reason: %s", err.Error())
	}

	err = checkAuth(cluster.Auth.Audience)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Prefix = fmt.Sprintf("--- Press %s to continue on %s ", color.GreenString("enter"), color.YellowString(cluster.Shorthand))
	s.Start()

	var input string
	_, err = fmt.Scanln(&input)
	s.Stop()

	if err != nil {
		if err.Error() == "unexpected newline" {
			return nil
		} else if err.Error() == "EOF" {
			fmt.Println("")
			return fmt.Errorf("command cancelled")
		} else {
			return err
		}
	}
	return nil
}

func displayCurrentContext(cmd *cobra.Command, args []string) error {
	skipConfirmation, err := cmd.Flags().GetBool("confirm")
	if err != nil {
		return err
	}

	if skipConfirmation {
		return nil
	}

	c, err := config.Load()
	if err != nil {
		return err
	}

	cluster, err := c.CurrentClusterConfig()
	if err != nil {
		return err
	}

	err = checkAuth(cluster.Auth.Audience)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "--- Running on %s\n", color.YellowString(cluster.Shorthand))

	return nil
}

func initErrorReporters() {
	report.ErrorReporters = []report.ErrorReporter{
		report.Sentry{
			Enable: viper.GetBool("TUBER_SENTRY_ENABLED"),
			Options: sentry.ClientOptions{
				Dsn:              viper.GetString("TUBER_SENTRY_DSN"),
				Environment:      viper.GetString("TUBER_SENTRY_ENVIRONMENT"),
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
	return appNameFlag
}

func fetchPodname() (string, error) {
	if pod != "" {
		return pod, nil
	}
	template := `{{range $k, $v := $.spec.selector.matchLabels}}{{$k}}={{$v}},{{end}}`
	l, err := k8s.Get("deployment", fetchWorkload(), appNameFlag, "-o", "go-template", "--template", template)
	if err != nil {
		return "", err
	}

	labels := strings.TrimSuffix(string(l), ",")

	jsonPath := fmt.Sprintf(`-o=jsonpath="%s"`, `{.items[0].metadata.name}`)
	podNameByte, err := k8s.GetCollection("pods", appNameFlag, "-l", labels, jsonPath)
	if err != nil {
		return "", err
	}
	return strings.Trim(string(podNameByte), "\""), nil
}

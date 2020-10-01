package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"tuber/pkg/core"
	"tuber/pkg/errorReporting"
	"tuber/pkg/k8s"

	"github.com/getsentry/sentry-go"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func clusterData() (*core.ClusterData, error) {
	defaultGateway := viper.GetString("cluster-default-gateway")
	defaultHost := viper.GetString("cluster-default-host")
	if defaultGateway == "" || defaultHost == "" {
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
	}

	data := &core.ClusterData{
		DefaultGateway: defaultGateway,
		DefaultHost:    defaultHost,
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
			return nil, err
		}
		return []byte(config.Data["credentials.json"]), nil
	}

	return creds, nil
}

type tuberConfig struct {
	Clusters []Cluster
}

// Cluster is a cluster
type Cluster struct {
	Name      string `yaml:"name"`
	Shorthand string `yaml:"shorthand"`
	URL       string `yaml:"url"`
}

func (c tuberConfig) CurrentClusterConfig() Cluster {
	name, err := k8s.CurrentCluster()
	if err != nil {
		return Cluster{}
	}

	return c.FindByName(name)
}

func (c tuberConfig) FindByShortName(name string) Cluster {
	for _, cl := range c.Clusters {
		if cl.Shorthand == name {
			return cl
		}
	}

	return Cluster{}
}

func (c tuberConfig) FindByName(name string) Cluster {
	for _, cl := range c.Clusters {
		if cl.Name == name {
			return cl
		}
	}

	return Cluster{}
}

func getTuberConfig() (*tuberConfig, error) {
	path, err := tuberConfigPath()
	if err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var t tuberConfig
	err = yaml.Unmarshal(raw, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func tuberConfigPath() (string, error) {
	dir, err := tuberConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "config.yaml"), nil
}

func tuberConfigDir() (string, error) {
	basePath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(basePath, "tuber"), nil
}

func promptCurrentContext(cmd *cobra.Command, args []string) error {
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
	fmt.Fprintf(os.Stderr, "Press ctrl+C to cancel, enter to continue...")
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
	cluster, err := k8s.CurrentCluster()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Running %s on %s", cmd.Name(), cluster)

	return nil
}

func errorReportingChannel(logger *zap.Logger) chan error {
	integrations := errorReporting.ErrorIntegrations{
		Reporters: []errorReporting.ErrorReporter{
			errorReporting.Sentry{
				Enable: viper.GetBool("sentry-enabled"),
				Options: sentry.ClientOptions{
					Dsn:              viper.GetString("sentry-dsn"),
					AttachStacktrace: true,
				},
			},
		},
	}
	errReports, err := errorReporting.StartWatching(integrations, logger)
	if err != nil {
		panic(err)
	}
	return errReports
}

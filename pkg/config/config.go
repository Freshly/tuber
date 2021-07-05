package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/freshly/tuber/pkg/k8s"
	"github.com/goccy/go-yaml"
)

type tuberConfig struct {
	ActiveClusterName string `yaml:"active_cluster_name"`
	ConfigSourceUrl   string `yaml:"config_source_url"`
	Clusters          []*cluster
}

// all pulled from the "APIs and Services - Credentials page"
// The backend resource powering IAP will have a web client with an ID
// A tuber client must also be created there, as a Desktop App. The client and secret are BOTH not actually secret for that type.
type auth struct {
	TuberDesktopClientID     string `yaml:"tuber_desktop_client_id"`
	TuberDesktopClientSecret string `yaml:"tuber_desktop_client_secret"`
	Audience                 string `yaml:"iap_backend_web_client_id"`
}

type cluster struct {
	Name      string `yaml:"name"`
	Shorthand string `yaml:"shorthand"`
	URL       string `yaml:"url"`
	Auth      *auth  `yaml:"auth"`
}

func (c *tuberConfig) SetActive(cl *cluster) error {
	c.ActiveClusterName = cl.Name
	err := Save(c)
	if err != nil {
		return fmt.Errorf("config invalid, please run 'tuber config'")
	}
	return nil
}

func (c *tuberConfig) CurrentClusterConfig() (*cluster, error) {
	k8sCheckErr := exec.Command("kubectl", "version", "--client").Run()
	k8sPresent := k8sCheckErr == nil

	if k8sPresent {
		kctlClusterName, err := k8s.CurrentCluster()
		if err != nil {
			return nil, fmt.Errorf("kubectl detected, but `kubectl config current-context` failed")
		}
		kctlClusterName = strings.Trim(kctlClusterName, "\r\n")

		k8sCluster, err := c.FindByName(kctlClusterName)
		if err != nil {
			return nil, fmt.Errorf("kubectl cluster %s not in config, please run 'tuber switch' or 'tuber config'", kctlClusterName)
		}

		var cl *cluster
		if c.ActiveClusterName != "" {
			cl, err = c.FindByName(c.ActiveClusterName)
			if err != nil {
				return nil, fmt.Errorf("active cluster not in config, please run 'tuber switch' or 'tuber config'")
			}
		}

		if kctlClusterName != c.ActiveClusterName {
			marshalErr := c.SetActive(k8sCluster)
			if marshalErr != nil {
				return nil, fmt.Errorf("config invalid, please run 'tuber config'")
			}
			return k8sCluster, nil
		}

		return cl, nil
	}

	if c.ActiveClusterName == "" {
		return nil, fmt.Errorf("active cluster not set, please run 'tuber switch'")
	}

	cl, err := c.FindByName(c.ActiveClusterName)
	if err != nil {
		return nil, fmt.Errorf("active cluster not in config, please run 'tuber switch' or 'tuber config'")
	}

	return cl, nil
}

func (c *tuberConfig) FindByShortName(name string) (*cluster, error) {
	if name == "" {
		return nil, fmt.Errorf("shorthand empty")
	}

	for _, cl := range c.Clusters {
		if cl.Shorthand == name {
			return cl, nil
		}
	}

	return nil, fmt.Errorf("cluster not found in config")
}

func (c *tuberConfig) FindByName(name string) (*cluster, error) {
	if name == "" {
		return nil, fmt.Errorf("name empty")
	}

	for _, cl := range c.Clusters {
		if cl.Name == name {
			return cl, nil
		}
	}

	return nil, fmt.Errorf("cluster not found in config")
}

func Load() (*tuberConfig, error) {
	path, err := Path()
	if err != nil {
		return nil, fmt.Errorf("tuber config not found, please run `tuber config`")
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tuber config not readable, please run `tuber config`")
	}

	var t tuberConfig
	err = yaml.Unmarshal(raw, &t)
	if err != nil {
		return nil, fmt.Errorf("tuber config invalid, please run `tuber config`")
	}

	var nonNilAuthClusters []*cluster
	for _, cl := range t.Clusters {
		if cl.Auth == nil {
			cl.Auth = &auth{}
		}
		nonNilAuthClusters = append(nonNilAuthClusters, cl)
	}

	t.Clusters = nonNilAuthClusters

	return &t, nil
}

func Save(config *tuberConfig) error {
	path, err := Path()
	if err != nil {
		return nil
	}

	out, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, os.ModePerm)
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "config.yaml"), nil
}

func Dir() (string, error) {
	basePath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(basePath, "tuber"), nil
}

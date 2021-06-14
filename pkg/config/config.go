package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/freshly/tuber/pkg/k8s"
	"github.com/goccy/go-yaml"
)

type tuberConfig struct {
	Clusters []Cluster
	Auth     Auth
}

// Auth is
type Auth struct {
	OAuthClientID string `yaml:"oauth_client_id"`
	OAuthSecret   string `yaml:"oauth_secret"`
}

// Cluster is a cluster
type Cluster struct {
	Name        string `yaml:"name"`
	Shorthand   string `yaml:"shorthand"`
	URL         string `yaml:"url"`
	IAPClientID string `yaml:"iap_client_id"`
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

	return &t, nil
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

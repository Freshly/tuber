package config

import (
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

func MustLoad() *tuberConfig {
	config, err := Load()

	if err != nil {
		panic(err)
	}

	return config
}

func Load() (*tuberConfig, error) {
	path, err := Path()
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

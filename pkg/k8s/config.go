package k8s

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

// configParser represents a part of kubectl's local config
type configParser struct {
	Users []struct {
		ClusterUser struct {
			Name     string `yaml:"name"`
			UserData struct {
				AuthProvider struct {
					Config struct {
						AccessToken string `yaml:"access-token"`
					} `yaml:"config"`
				} `yaml:"auth-provider"`
			} `yaml:"user"`
		}
	} `yaml:"users"`
}

// ClusterConfig returns config for a cluster
type ClusterConfig struct {
	Name        string
	AccessToken string
}

// GetConfig returns `kubectl config view`
func GetConfig() (*ClusterConfig, error) {
	var config configParser

	out, err := kubectl([]string{"config", "view", "--raw"}...)
	if err != nil {
		return &ClusterConfig{}, err
	}

	yaml.Unmarshal(out, &config)

	clusterName, err := CurrentCluster()
	if err != nil {
		return &ClusterConfig{}, err
	}

	for _, cnf := range config.Users {
		if cnf.ClusterUser.Name == clusterName {
			return &ClusterConfig{
				Name:        cnf.ClusterUser.Name,
				AccessToken: cnf.ClusterUser.UserData.AuthProvider.Config.AccessToken,
			}, nil
		}
	}

	return &ClusterConfig{}, fmt.Errorf("no config found for current cluster")
}

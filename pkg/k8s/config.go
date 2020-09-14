package k8s

import (
	"encoding/json"
	"fmt"
)

// configParser represents a part of kubectl's local config
type configParser struct {
	Users []struct {
		Name        string `json:"name"`
		ClusterUser struct {
			UserData struct {
				AuthProvider struct {
					AccessToken string `json:"access-token"`
				} `json:"config"`
			} `json:"auth-provider"`
		} `json:"user"`
	} `json:"users"`
}

// ClusterConfig returns config for a cluster
type ClusterConfig struct {
	Name        string
	AccessToken string
}

// GetConfig returns `kubectl config view`
func GetConfig() (*ClusterConfig, error) {
	var config configParser

	out, err := kubectl([]string{"config", "view", "-o", "json"}...)
	if err != nil {
		return &ClusterConfig{}, err
	}

	json.Unmarshal(out, &config)

	clusterName, err := CurrentCluster()
	if err != nil {
		return &ClusterConfig{}, err
	}

	for _, cnf := range config.Users {
		if cnf.Name == clusterName {
			return &ClusterConfig{
				Name:        cnf.Name,
				AccessToken: cnf.ClusterUser.UserData.AuthProvider.AccessToken,
			}, nil
		}
	}

	return &ClusterConfig{}, fmt.Errorf("no config found for current cluster")
}

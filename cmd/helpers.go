package cmd

import (
	"io/ioutil"
	"tuber/pkg/core"

	"github.com/goccy/go-yaml"
	"github.com/spf13/viper"
)

func clusterData() (data *core.ClusterData) {
	return &core.ClusterData{
		DefaultGateway: viper.GetString("cluster-default-gateway"),
		DefaultHost:    viper.GetString("cluster-default-host"),
	}
}

func credentials() (creds []byte, err error) {
	viper.SetDefault("credentials-path", "/etc/tuber-credentials/credentials.json")
	credentialsPath := viper.GetString("credentials-path")
	creds, err = ioutil.ReadFile(credentialsPath)
	return
}

func getTuberrc() (*tuberrc, error) {
	path := viper.GetString("tuberrc-path")
	if path == "" {
		return nil, nil
	}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var t tuberrc
	err = yaml.Unmarshal(raw, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

type tuberrc struct {
	Clusters map[string]string
}

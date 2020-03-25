package cmd

import (
	"io/ioutil"
	"tuber/pkg/core"

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

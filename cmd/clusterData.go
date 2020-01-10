package cmd

import (
	"tuber/pkg/core"

	"github.com/spf13/viper"
)

func clusterData() (data *core.ClusterData) {
	return &core.ClusterData{
		DefaultGateway: viper.GetString("cluster-default-gateway"),
		Hostname:       viper.GetString("cluster-hostname"),
	}
}

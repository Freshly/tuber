package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tuber",
		Short: "CLI to manage containerized applications on GKE",
	}
)

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug")
	rootCmd.PersistentFlags().BoolP("confirm", "y", false, "automatic yes to prompts")
	_ = viper.BindPFlag("TUBER_DEBUG", rootCmd.PersistentFlags().Lookup("debug"))
}

func createLogger() (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if viper.GetBool("TUBER_DEBUG") {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	return logger, err
}

// Execute executes
func Execute() error {
	return rootCmd.Execute()
}

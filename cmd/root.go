package cmd

import (
	"strings"

	"github.com/joho/godotenv"
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
	// Environment variables prefixed with `TUBER_` are immediately available
	// to Viper with '-' substitution. E.g., `TUBER_DEBUG=true` is available as
	// `viper.GetBool("debug")`
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TUBER")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug")
	rootCmd.PersistentFlags().BoolP("confirm", "y", false, "automatic yes to prompts")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func createLogger() (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if viper.GetBool("debug") {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	return logger, err
}

// Execute executes
func Execute() error {
	godotenv.Load()

	return rootCmd.Execute()
}

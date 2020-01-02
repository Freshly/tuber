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
		Short: "",
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
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func createLogger() (logger *zap.Logger) {
	if viper.GetBool("debug") {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	return
}

// Execute executes
func Execute() error {
	godotenv.Load()

	return rootCmd.Execute()
}

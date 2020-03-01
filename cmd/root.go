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

	// The appName flag binds to this variable
	appName string
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

func createLogger() (logger *zap.Logger, err error) {
	if viper.GetBool("debug") {
		logger, err = zap.NewDevelopment()
		if err != nil {
			return
		}
	} else {
		logger, err = zap.NewProduction()
		if err != nil {
			return
		}
	}

	return
}

// Execute executes
func Execute() error {
	godotenv.Load()

	return rootCmd.Execute()
}

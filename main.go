package main

import (
	"log"

	"github.com/freshly/tuber/cmd"
	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("prefix", "/tuber")
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"

	"github.com/freshly/tuber/cmd"
)

var (
	version = "dev"
)

func main() {
	cmd.Version = version
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

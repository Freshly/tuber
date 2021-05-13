package main

import (
	"log"

	"github.com/freshly/tuber/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

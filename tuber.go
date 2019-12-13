package main

import (
	"fmt"
	"log"
	"os"

	"tuber/pkg/apply"
	"tuber/pkg/yaml_retrieval"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	name := os.Getenv("IMAGE_NAME")
	tag := os.Getenv("IMAGE_TAG")

	yamls, err := yaml_retrieval.RetrieveAll(name, tag)

	if err != nil {
		log.Fatal(err)
	}

	out, err := apply.Apply(yamls)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)
}

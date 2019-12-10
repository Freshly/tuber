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

	image := yaml_retrieval.ImageInfo{ Name: os.Getenv("IMAGE_NAME"), Tag: os.Getenv("IMAGE_TAG") }

	retriever := yaml_retrieval.Retriever{Image: image}
	yamls, err := retriever.RetrieveAll()

	if err != nil {
		log.Fatal(err)
	}

	out, err := apply.Apply(yamls)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", out)
}

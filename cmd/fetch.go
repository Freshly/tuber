package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"tuber/pkg/yamldownloader"
)

func init() {
	rootCmd.AddCommand(fetchCmd)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch display yamls",

	Run: fetch,
}

func fetch(cmd *cobra.Command, args []string) {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var registry = yamldownloader.NewGoogleRegistry(os.Getenv("GCLOUD_TOKEN"))
	repository, err := registry.GetRepository(os.Getenv("IMAGE_NAME"), "pull")

	if err != nil {
		log.Fatal(err)
	}

	yamls, err := repository.FindLayer(os.Getenv("IMAGE_TAG"))

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", yamls)
}

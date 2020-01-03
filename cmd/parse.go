package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "apply local yams",
	Run:   parse,
}

func init() {
	rootCmd.AddCommand(parseCmd)
}

func parse(cmd *cobra.Command, args []string) {
	data, readErr := ioutil.ReadFile(".tuber/deploy.json")
	if readErr != nil {
		fmt.Println(readErr)
		return
	}

	query, err := gojq.Parse(`.spec.template.spec.containers.[0].image = "kjladfslfdjkaslfdasasfdlsafdfdsa"`)
	if err != nil {
		log.Fatalln(err)
	}

	var obj interface{}
	json.Unmarshal(data, &obj)

	iter := query.Run(obj.(interface{}).(map[string]interface{}))

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}
		fmt.Printf("%#v\n", v)
	}

	if err != nil {
		log.Fatal(err)
	}
}

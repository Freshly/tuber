package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var adminserverCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "adminserver",
	Short:        "starts the admin http server for review apps and maybe other stuff who knows",
	RunE:         adminserver,
}

func adminserver(cmd *cobra.Command, args []string) error {
	http.HandleFunc("/tuber", helloServer)
	http.ListenAndServe(":3000", nil)
	return nil
}

func helloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>hello im tuber</h1>")
}

func init() {
	rootCmd.AddCommand(adminserverCmd)
}

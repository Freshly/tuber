package adminserver

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"tuber/pkg/reviewapps"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var dashboardPage = `<p>create a review app from this branch:</p>
<form method="POST">
<span>Source app name:</span><input type="text" name="appname">
<br>
<span>Branch name:</span><input type="text" name="branch">
<input type="submit" value="Submit">
</form>
<br>
{{if .Result}}
{{.Result}}
{{end}}
`

var projectName string
var creds []byte
var logger *zap.Logger

func Start(p string, c []byte, l *zap.Logger) {
	projectName = p
	creds = c
	logger = l
	http.HandleFunc("/tuber", dashboard)
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("").Parse(dashboardPage))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	branch := r.FormValue("branch")
	appName := r.FormValue("appname")
	reviewAppName, err := reviewapps.CreateReviewApp(context.Background(), &zap.Logger{}, branch, appName, creds, projectName)
	var result string
	if err == nil {
		result = fmt.Sprintf("https://%s.%s/", reviewAppName, viper.GetString("cluster-default-host"))
	} else {
		result = err.Error()
	}

	tmpl.Execute(w, struct{ Result string }{result})
}

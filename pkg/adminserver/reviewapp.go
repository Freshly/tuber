package adminserver

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
	"tuber/pkg/core"
	"tuber/pkg/k8s"
	"tuber/pkg/reviewapps"

	"github.com/gorilla/mux"
	"google.golang.org/api/cloudbuild/v1"
)

func reviewAppHandler() reviewApp {
	return reviewApp{
		nameKey:          "reviewAppName",
		sourceAppNameKey: sourceAppHandler().nameKey,
		pathPrefix:       "reviewapps",
		raw: `<head><title>Tuber Admin: {{.Name}}</title></head>
{{if .Error}}
	<p>{{.Error}}</p>
{{else}}
	<a href="/tuber"><- back to all apps</a>
	<h1>{{.Name}}</h1>
	<p>Available at - <a href="{{.Link}}">{{.Link}}</a> - if it uses your cluster's default hostname. Otherwise extrapolate from app name: {{.Name}}</p>
	<p>created from <a href="../../{{.SourceAppName}}">{{.SourceAppName}}</a></p>
	<h2>Builds</h2>
	<table>{{range .Builds}}
		<tr>
			<td><a href="{{.Link}}">{{.StartTime}}</a></td>
			<td>{{.Status}}</td>
		</tr>
	{{end}}</table>
{{end}}
`,
	}
}

type reviewApp struct {
	raw                 string
	reviewAppsEnabled   bool
	sourceAppNameKey    string
	nameKey             string
	cloudbuildClient    *cloudbuild.Service
	triggersProjectName string
	clusterDefaultHost  string
	pathPrefix          string
	data                struct {
		SourceAppName string
		Name          string
		Branch        string
		DatadogLink   string
		Builds        []Build
		Error         string
		Link          string
	}
}

type Build struct {
	Status    string
	Link      string
	StartTime string
}

func (r reviewApp) path() string {
	return sourceAppHandler().path() + fmt.Sprintf("/%s/{%s}", r.pathPrefix, r.nameKey)
}

func (r reviewApp) setup(reviewAppsEnabled bool, cloudbuildClient *cloudbuild.Service, triggersProjectName string, clusterDefaultHost string) handler {
	r.reviewAppsEnabled = reviewAppsEnabled
	r.cloudbuildClient = cloudbuildClient
	r.triggersProjectName = triggersProjectName
	r.clusterDefaultHost = clusterDefaultHost
	return r
}

func (r reviewApp) handle(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	sourceAppName := vars[r.sourceAppNameKey]
	reviewAppName := vars[r.nameKey]
	tmpl := template.Must(template.New("").Parse(r.raw))
	data := &r.data
	defer tmpl.Execute(w, data)

	data.Name = reviewAppName
	data.SourceAppName = sourceAppName
	data.Link = dumbLink(reviewAppName, r.clusterDefaultHost)

	if !r.reviewAppsEnabled {
		w.WriteHeader(http.StatusNotFound)
		data.Error = "review apps are not enabled on this cluster"
		return
	}

	allReviewApps, err := core.TuberReviewApps()
	if err != nil {
		data.Error = err.Error()
		return
	}

	reviewApp, err := allReviewApps.FindApp(reviewAppName)
	if err != nil {
		data.Error = err.Error()
		return
	}
	data.Branch = reviewApp.Tag

	config, err := k8s.GetConfigResource(reviewapps.TuberReviewTriggersConfig, "tuber", "configmap")
	triggerId := config.Data[reviewAppName]
	if triggerId == "" {
		data.Error = "trigger is untracked or it doesnt exist"
		return
	}

	buildsResponse, err := cloudbuild.NewProjectsBuildsService(r.cloudbuildClient).List(r.triggersProjectName).Filter(fmt.Sprintf(`trigger_id="%s"`, triggerId)).Do()
	if err != nil {
		data.Error = err.Error()
		return
	}
	var builds []Build
	for _, build := range buildsResponse.Builds {
		var startTime string
		if build.StartTime != "" {
			parsed, timeErr := time.Parse(time.RFC3339, build.StartTime)
			if timeErr == nil {
				startTime = parsed.String()
			}
		}
		builds = append(builds, Build{Status: strings.Title(build.Status), Link: build.LogUrl, StartTime: startTime})
	}
	data.Builds = builds
}

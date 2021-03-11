package adminserver

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"tuber/pkg/core"

	"github.com/gorilla/mux"
)

func sourceAppHandler() sourceApp {
	return sourceApp{
		nameKey:  "sourceAppName",
		appsPath: "apps",
		raw: `<head><title>Tuber Admin: {{.Name}}</title></head>
<a href="/tuber"><- back to all apps</a>
<h1>{{.Name}}</h1>
<p>Available at - <a href="{{.Link}}">{{.Link}}</a> - if it uses your cluster's default hostname. Otherwise extrapolate from app name: {{.Name}}</p>
{{if .ReviewAppsEnabled}}
	<br>
	<h2>Create a review app</h2>
	<form method="POST" action="{{.ReviewAppCreationPath}}">
		<input type="hidden"name="appname" value="{{.Name}}">
		<span>Branch name:</span><input type="text" name="branch">
		<input type="submit" value="Submit">
	</form>
{{end}}
{{if .Error}}
	<p>{{.Error}}</p>
	<br>
{{end}}
{{if .ReviewAppsEnabled}}
	<br>
	<h2>Review Apps</h2>
	<table>
	<th>App</th>
	<th>Branch</th>
	{{range .ReviewApps}}
		<tr>
			<td><a href="{{$.ReviewAppShowPath}}/{{.Name}}">{{.Name}}</a></td>
			<td>{{.Branch}}</td>
		</tr>
	{{end}}</table>
{{end}}
`,
	}
}

type sourceApp struct {
	raw                string
	reviewAppsEnabled  bool
	nameKey            string
	clusterDefaultHost string
	appsPath           string
	data               struct {
		Name                  string
		DatadogLink           string
		ReviewAppsEnabled     bool
		ReviewApps            []ReviewApp
		Link                  string
		Error                 string
		ReviewAppCreationPath string
		ReviewAppShowPath     string
	}
}

type ReviewApp struct {
	Name   string
	Branch string
}

func (s sourceApp) path() string { return fmt.Sprintf("/apps/{%s}", s.nameKey) }

func (s sourceApp) setup(reviewAppsEnabled bool, clusterDefaultHost string) handler {
	s.reviewAppsEnabled = reviewAppsEnabled
	s.clusterDefaultHost = clusterDefaultHost
	return s
}

func (s sourceApp) handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars[s.nameKey]
	tmpl := template.Must(template.New("").Parse(s.raw))
	data := &s.data
	data.Name = name
	data.ReviewAppsEnabled = s.reviewAppsEnabled
	data.Link = dumbLink(name, s.clusterDefaultHost)

	data.ReviewAppCreationPath = r.URL.Path + "/" + createReviewAppHandler().creationPath
	data.ReviewAppShowPath = r.URL.Path + "/" + reviewAppHandler().pathPrefix
	reviewAppCreationErr := r.URL.Query().Get(createReviewAppHandler().errorParam)

	defer tmpl.Execute(w, data)

	if reviewAppCreationErr != "" {
		data.Error = reviewAppCreationErr
	}

	if s.reviewAppsEnabled {
		allReviewApps, err := core.TuberReviewApps()
		if err != nil {
			data.Error = err.Error()
			return
		}

		sourceApps, err := core.TuberSourceApps()
		if err != nil {
			data.Error = err.Error()
			return
		}

		sourceApp, err := sourceApps.FindApp(name)
		if err != nil {
			data.Error = err.Error()
			return
		}

		var reviewApps []ReviewApp
		for _, reviewApp := range allReviewApps {
			if sourceApp.Repo == reviewApp.Repo {
				reviewApps = append(reviewApps, ReviewApp{Name: reviewApp.Name, Branch: reviewApp.Tag})
			}
		}
		sort.Slice(reviewApps, func(i, j int) bool {
			return reviewApps[i].Name < reviewApps[j].Name
		})

		data.ReviewApps = reviewApps
	}
}

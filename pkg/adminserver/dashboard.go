package adminserver

import (
	"html/template"
	"net/http"
	"sort"
	"tuber/pkg/core"
)

func dashboardHandler() dashboard {
	return dashboard{
		routePath: "/",
		raw: `<head><title>Tuber Admin</title></head>
<h1>potate</h1>
{{if .Error}}
	<p>{{.Error}}</p>
{{else}}
	<table>{{range .SourceApps}}
		<tr>
			<td><a href="{{$.AppPath}}/{{.Name}}">{{.Name}}</a></td>
			<td>{{.Tag}}</td>
		</tr>
	{{end}}</table>
{{end}}
`,
	}
}

type dashboard struct {
	raw       string
	routePath string
	data      struct {
		AppPath    string
		SourceApps []SourceApp
		Error      string
	}
}

func (d dashboard) path() string { return d.routePath }

func (d dashboard) setup() handler { return d }

type SourceApp struct {
	Name string
	Tag  string
}

func (d dashboard) handle(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("").Parse(d.raw))
	data := &d.data
	defer tmpl.Execute(w, data)
	data.AppPath = sourceAppHandler().appsPath

	apps, err := core.TuberSourceApps()
	if err != nil {
		data.Error = err.Error()
		return
	}

	var sourceApps []SourceApp
	for _, app := range apps {
		sourceApps = append(sourceApps, SourceApp{Name: app.Name, Tag: app.ImageTag})
	}
	sort.Slice(sourceApps, func(i, j int) bool {
		return sourceApps[i].Name < sourceApps[j].Name
	})

	data.SourceApps = sourceApps
}

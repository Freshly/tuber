package core

import (
	"bytes"
	"text/template"
	"tuber/pkg/k8s"
)

// ApplyTemplate interpolates and applies a yaml to a given namespace
func ApplyTemplate(namespace string, templateString string, data map[string]string) error {
	interpolated, err := interpolate(templateString, data)
	if err != nil {
		return err
	}
	return k8s.Apply(interpolated, namespace)
}

func interpolate(templateString string, data map[string]string) (interpolated []byte, err error) {
	tpl, err := template.New("").Parse(templateString)

	if err != nil {
		return
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)

	if err != nil {
		return
	}

	interpolated = buf.Bytes()
	return
}

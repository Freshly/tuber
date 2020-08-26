package core

import (
	"io/ioutil"
	"os"
	data "tuber/data/tuberapps"
)

const tuberConfigPath = ".tuber"

// InitTuberApp creates a bunch of yamls for you
func InitTuberApp(appName string, routePrefix string, withIstio bool, serviceType string) error {
	err := createTuberDirectory()
	if err != nil {
		return err
	}

	err = createDeploymentYAML(appName)
	if err != nil {
		return err
	}

	err = modDockerFile()
	if err != nil {
		return err
	}

	if !withIstio {
		return nil
	}

	err = createServiceYAML(appName, serviceType)
	if err != nil {
		return err
	}

	return createVirtualServiceYAML(appName, routePrefix)
}

func createTuberDirectory() error {
	err := os.Mkdir(".tuber", os.ModePerm)
	if os.IsExist(err) {
		return nil
	}
	return err
}

func createDeploymentYAML(appName string) error {
	templateData := map[string]string{
		"appName": appName,
	}

	return writeYAML(data.Deployment, templateData)
}

func createServiceYAML(appName string, serviceType string) error {
	templateData := map[string]string{
		"appName":     appName,
		"serviceType": serviceType,
	}

	return writeYAML(data.Service, templateData)
}

func createVirtualServiceYAML(appName string, routePrefix string) error {
	templateData := map[string]string{
		"appName":     appName,
		"routePrefix": routePrefix,
	}

	return writeYAML(data.Virtualservice, templateData)
}

func modDockerFile() error {
	f, err := os.OpenFile("./Dockerfile", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("COPY .tuber /.tuber"); err != nil {
		return err
	}

	return nil
}

func writeYAML(app data.TuberYaml, templateData map[string]string) error {
	interpolated, err := interpolate(string(app.Contents), templateData)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(tuberConfigPath+"/"+app.Filename, interpolated, 0644)
}

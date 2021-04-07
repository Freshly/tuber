package k8s

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/goccy/go-yaml"
)

// CreateTuberCredentials creates a secret based on the contents of a file
func CreateTuberCredentials(path string, namespace string) (err error) {
	dat, err := ioutil.ReadFile(path)

	if err != nil {
		return
	}

	str := string(dat)
	filename := filepath.Base(path)
	data := map[string]string{filename: str}
	meta := k8sMetadata{
		Name:      fmt.Sprintf("%s-%s", namespace, filename),
		Namespace: namespace,
	}

	config := k8sConfigResource{
		APIVersion: "v1",
		Kind:       "Secret",
		Type:       "Opaque",
		StringData: data,
		Metadata:   meta,
	}

	var jsondata []byte
	jsondata, err = json.Marshal(config)

	if err != nil {
		return
	}

	return Apply(jsondata, namespace)
}

func GetSecret(namespace string, secretName string) (*ConfigResource, error) {
	config, err := GetConfigResource(secretName, namespace, "Secret")
	if err != nil {
		return nil, err
	}

	for k, v := range config.Data {
		decoded, decodeErr := base64.StdEncoding.DecodeString(v)
		if decodeErr != nil {
			return nil, decodeErr
		}
		config.Data[k] = string(decoded)
	}

	return config, nil
}

// CreateEnvFromFile replaces an apps env with data in a local file
func CreateEnvFromFile(name string, path string) error {
	var fileBody []byte
	var err error

	if path == "-" {
		fileBody, err = ioutil.ReadAll(bufio.NewReader(os.Stdin))
	} else {
		fileBody, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return err
	}

	config, err := GetConfigResource(name+"-env", name, "Secret")
	if err != nil {
		return err
	}

	config.Data = processData(fileBody)
	return config.Save(name)
}

func processData(dataIn []byte) map[string]string {
	if dataIn == nil {
		return nil
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(dataIn, &data); err != nil {
		return nil
	}

	// look for a value of ${SOME_LOCAL_ENV_VAR}
	// works with env vars that have hyphens or no separators
	var envVar = regexp.MustCompile(`\$\{(.*)\}`)

	var out = make(map[string]string, len(data))
	for k, v := range data {
		val := fmt.Sprint(v)
		found := envVar.FindStringSubmatch(val)
		if len(found) > 1 {
			val = os.Getenv(found[1])
		}
		out[k] = base64.StdEncoding.EncodeToString([]byte(val))
	}

	return out
}

// PatchSecret gets, patches, and saves a secret
func PatchSecret(mapName string, namespace string, key string, value string) (err error) {
	config, err := GetConfigResource(mapName, namespace, "Secret")

	if err != nil {
		return
	}

	value = base64.StdEncoding.EncodeToString([]byte(value))

	if config.Data == nil {
		config.Data = map[string]string{key: value}
	} else {
		config.Data[key] = value
	}

	return config.Save(namespace)
}

// RemoveSecretEntry removes an entry, from a secret
func RemoveSecretEntry(mapName string, namespace string, key string) (err error) {
	config, err := GetConfigResource(mapName, namespace, "Secret")

	if err != nil {
		return
	}

	delete(config.Data, key)

	return config.Save(namespace)
}

// CreateEnv creates a Secret for a new TuberApp, to store env vars
func CreateEnv(appName string) error {
	return Create(appName, "secret", "generic", appName+"-env")
}

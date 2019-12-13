package yaml_retrieval

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"net/http"
	"os"
	"tuber/pkg/util"
)

type image struct {
	name string
	tag string
}

type authorizedImage struct {
	image
	token string
}

type registryToken struct {
	Token string `json:"token"`
}

// RetrieveAll returns all Yamls for the env-configured registry
func RetrieveAll(name string, tag string) (yamls []util.Yaml, err error) {
	image, err := image { name: name, tag: tag }.authorize()
	if err != nil { return }

	manifest, err := image.getManifest()
	if err != nil { return }

	yamls, err = manifest.downloadYamls()
	return
}

func (i authorizedImage) getManifest() (m manifest, err error) {
	requestURL := fmt.Sprintf(
		"%s/v2/%s/manifests/%s",
		os.Getenv("REGISTRY_BASE"),
		i.name,
		i.tag,
	)

	client := &http.Client{}

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", i.token))
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	res, err := client.Do(req)
	fmt.Println("------response:")
	fmt.Println(res)

	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	manifest := &manifest { image: i }

	fmt.Println("------body:")
	fmt.Println(body)
	err = json.Unmarshal(body, manifest)

	if err != nil {
		return
	}

	spew.Dump(manifest)
	return *manifest, nil
}

func (i image) authorize() (image authorizedImage, err error) {
	requestURL := fmt.Sprintf(
		"%s/v2/token?scope=repository:%s:pull",
		os.Getenv("AUTH_BASE"),
		i.name,
	)

	client := &http.Client{}

	req, err := http.NewRequest("GET", requestURL, nil)

	if err != nil {
		return
	}

	req.SetBasicAuth("_token", os.Getenv("GCLOUD_TOKEN"))
	res, err := client.Do(req)

	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	tokenResponse := new(registryToken)
	err = json.Unmarshal(body, &tokenResponse)

	if err != nil {
		return
	}

	if tokenResponse.Token == "" {
		err = fmt.Errorf("no token")
		return
	}

	spew.Dump(tokenResponse.Token)

	image = authorizedImage { image: i, token: tokenResponse.Token }
	return
}
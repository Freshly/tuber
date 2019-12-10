package yaml_retrieval

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"tuber/pkg/util"
)

type notTuberLayerError struct {
	message string
}

type manifest struct {
	Layers []layer `json:"layers"`
}

type layer struct {
	Digest string `json:"digest"`
	Size   int32  `json:"size"`
}

func (e *notTuberLayerError) Error() string { return e.message }

func (r Retriever) getLayers() ([]layer, error) {
	requestURL := fmt.Sprintf(
		"%s/v2/%s/manifests/%s",
		os.Getenv("REGISTRY_BASE"),
		r.Image.Name,
		r.Image.Tag,
	)

	client := &http.Client{}

	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.AuthResponse.Token))
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var obj = new(manifest)
	err = json.Unmarshal(body, &obj)

	if err != nil {
		return nil, err
	}

	spew.Dump(obj)
	return obj.Layers, nil
}

func (r Retriever) downloadLayer(layerObj *layer) ([]util.Yaml, error) {
	layer := layerObj.Digest

	requestURL := fmt.Sprintf(
		"%s/v2/%s/blobs/%s",
		os.Getenv("REGISTRY_BASE"),
		r.Image.Name,
		layer,
	)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.AuthResponse.Token))

	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	gzipped, _ := gzip.NewReader(res.Body)
	archive := tar.NewReader(gzipped)
	var yamls []util.Yaml

	for {
		header, err := archive.Next()

		if err == io.EOF {
			break // End of archive
		}

		if err != nil {
			return nil, err
		}

		if !strings.HasPrefix(header.Name, ".tuber") {
			return nil, &notTuberLayerError{"contains stuff other than .tuber"}
		}

		if !strings.HasSuffix(header.Name, ".yaml") {
			continue
		}

		bytes, _ := ioutil.ReadAll(archive)

		var yaml util.Yaml
		yaml.Filename = header.Name
		yaml.Content = string(bytes)

		yamls = append(yamls, yaml)
	}

	return yamls, nil
}
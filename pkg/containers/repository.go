package containers

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"tuber/pkg/gcloud"
)

type RepositoryLocation struct {
	Host string
	Path string
	Tag  string
}

const megabyte = 1_000_000
const maxSize = megabyte * 1

type layer struct {
	Digest string `json:"digest"`
	Size   int32  `json:"size"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a apiError) Error() string {
	return a.Message
}

type manifest struct {
	Digest string
	Layers []layer    `json:"layers"`
	Errors []apiError `json:"errors"`
}

type repository struct {
	registry *registry
	path     string
	token    string
}

// GetTuberLayer downloads yamls for an image
func GetTuberLayer(location RepositoryLocation, sha string, creds []byte) (prereleaseYamls []string, releaseYamls []string, err error) {
	authToken, err := gcloud.GetAccessToken(creds)
	if err != nil {
		return
	}

	reg := newRegistry(location.Host, authToken)
	repo, err := reg.getRepository(location.Path)
	if err != nil {
		return
	}

	prereleaseYamls, releaseYamls, err = repo.findLayer(sha)
	return
}

func GetLatestSHA(location RepositoryLocation, creds []byte) (sha string, err error) {
	authToken, err := gcloud.GetAccessToken(creds)

	if err != nil {
		return
	}

	reg := newRegistry(location.Host, authToken)
	repo, err := reg.getRepository(location.Path)

	if err != nil {
		return
	}

	m, err := repo.getManifest(location.Tag)
	if err != nil {
		return
	}

	sha = m.Digest
	return
}

func (r *repository) getManifest(tag string) (m manifest, err error) {
	requestURL := fmt.Sprintf(
		"%s/v2/%s/manifests/%s",
		r.registry.baseURL,
		r.path,
		tag,
	)

	client := &http.Client{}

	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.token))
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	res, err := client.Do(req)

	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return
	}

	if len(res.Header["Docker-Content-Digest"]) == 0 {
		err = &InvalidRegistryResponse{StatusCode: res.StatusCode, Headers: res.Header}
		return
	}

	digest := res.Header["Docker-Content-Digest"][0]

	m = manifest{Digest: digest}
	err = json.Unmarshal(body, &m)

	if err != nil {
		return
	}

	if len(m.Errors) > 0 {
		err = m.Errors[0]
		return
	}

	return
}

func (r *repository) getLayers(tag string) (layers []layer, err error) {
	manifest, err := r.getManifest(tag)

	if err != nil {
		return
	}

	return manifest.Layers, nil
}

func (r *repository) downloadLayer(layerObj *layer) (prereleaseYamls []string, releaseYamls []string, err error) {
	layer := layerObj.Digest

	requestURL := fmt.Sprintf(
		"%s/v2/%s/blobs/%s",
		r.registry.baseURL,
		r.path,
		layer,
	)

	client := &http.Client{}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.token))

	res, err := client.Do(req)

	if err != nil {
		return
	}

	prereleaseYamls, releaseYamls, err = convertResponse(res)
	return
}

func convertResponse(response *http.Response) (prereleaseYamls []string, releaseYamls []string, err error) {
	gzipped, err := gzip.NewReader(response.Body)

	if err != nil {
		return
	}

	archive := tar.NewReader(gzipped)

	for {
		var header *tar.Header
		header, err = archive.Next()

		if err == io.EOF {
			err = nil
			return
		}

		if err != nil {
			return
		}

		if !strings.HasPrefix(header.Name, ".tuber") {
			continue
		}

		if !strings.HasSuffix(header.Name, ".yaml") {
			continue
		}

		var bytes []byte
		bytes, err = ioutil.ReadAll(archive)

		if err != nil {
			return
		}

		if strings.HasPrefix(header.Name, ".tuber/prerelease/") {
			prereleaseYamls = append(prereleaseYamls, string(bytes))
		} else {
			releaseYamls = append(releaseYamls, string(bytes))
		}
	}
}

// findLayer finds the .tuber layer containing deploy info for Tuber
func (r *repository) findLayer(tag string) (prereleaseYamls []string, releaseYamls []string, err error) {
	layers, err := r.getLayers(tag)

	if err != nil {
		return
	}

	for _, layer := range layers {
		if layer.Size > maxSize {
			continue
		}

		prereleaseYamls, releaseYamls, err = r.downloadLayer(&layer)

		if err != nil {
			return nil, nil, err
		}

		if len(releaseYamls) != 0 {
			return prereleaseYamls, releaseYamls, nil
		}
	}

	err = fmt.Errorf("no tuber layer found")
	return
}

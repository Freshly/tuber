package gcr

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const megabyte = 1_000_000
const maxSize = megabyte * 1

type AppYamls struct {
	Prerelease  []string
	Release     []string
	PostRelease []string
	Tags        []string
}

// GetTuberLayer downloads yamls for an image
func GetTuberLayer(tagOrDigest string, creds []byte) (*AppYamls, error) {
	ref, err := name.ParseReference(tagOrDigest)
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(ref, remote.WithAuth(google.NewJSONKeyAuthenticator(string(creds))))
	if err != nil {
		return nil, err
	}
	layers, err := img.Layers()
	if err != nil {
		return nil, err
	}
	yamls, err := getTuberYamls(layers)
	if err != nil {
		return nil, err
	}

	repoImages, err := google.List(ref.Context(), google.WithAuth(google.NewJSONKeyAuthenticator(string(creds))))
	if err != nil {
		return nil, err
	}

	digest, err := img.Digest()
	if err != nil {
		return nil, err
	}

	yamls.Tags = repoImages.Manifests[digest.String()].Tags
	return yamls, nil
}

func getTuberYamls(layers []v1.Layer) (*AppYamls, error) {
	var tuberYamls *AppYamls
	for i := len(layers) - 1; i >= 0; i-- {
		size, err := layers[i].Size()
		if size > maxSize {
			continue
		}

		yamls, err := findTuberYamls(layers[i])
		if err != nil {
			return nil, err
		}
		if yamls != nil {
			tuberYamls = yamls
			break
		}
	}
	return tuberYamls, nil
}

func findTuberYamls(layer v1.Layer) (*AppYamls, error) {
	var yamls AppYamls
	uncompressed, err := layer.Uncompressed()
	if err != nil {
		return nil, err
	}
	archive := tar.NewReader(uncompressed)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return &yamls, nil
		}

		if err != nil {
			return nil, err
		}

		fileName := header.Name

		if strings.HasPrefix(fileName, ".tuber/") && strings.HasSuffix(fileName, ".yaml") {
			var raw []byte
			raw, err = ioutil.ReadAll(archive)
			if err != nil {
				return nil, err
			}
			if strings.HasPrefix(fileName, ".tuber/prerelease/") {
				yamls.Prerelease = append(yamls.Prerelease, string(raw))
			} else if strings.HasPrefix(fileName, ".tuber/postrelease/") {
				yamls.PostRelease = append(yamls.PostRelease, string(raw))
			} else {
				yamls.Release = append(yamls.Release, string(raw))
			}
		}
	}
}

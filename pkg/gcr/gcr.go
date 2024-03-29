package gcr

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"go.uber.org/zap"
)

const megabyte = 1_000_000
const maxSize = megabyte * 1

func DigestFromTag(tag string, creds []byte) (string, error) {
	ref, err := name.ParseReference(tag)
	if err != nil {
		return "", err
	}

	img, err := remote.Image(ref, remote.WithAuth(google.NewJSONKeyAuthenticator(string(creds))))
	if err != nil {
		return "", err
	}

	digest, err := img.Digest()
	if err != nil {
		return "", err
	}

	return ref.Context().Digest(digest.String()).String(), nil
}

func SwapTags(imageTag string, tag string) (string, error) {
	ref, err := name.ParseReference(imageTag)
	if err != nil {
		return "", err
	}
	newTag := ref.Context().Tag(tag).String()
	_, err = name.ParseReference(newTag)
	if err != nil {
		return "", err
	}
	return newTag, nil
}

// Tag = Docker Tag = branchName
// Ref = Full refernence to image = gcr.io/freshly-docker/appName:branchName
func TagFromRef(imageTag string) (string, error) {
	ref, err := name.ParseReference(imageTag)
	if err != nil {
		return "", err
	}
	return ref.Identifier(), nil
}

type AppYamls struct {
	Prerelease  []string
	Release     []string
	PostRelease []string
	Tags        []string
}

// GetTuberLayer downloads yamls for an image
func GetTuberLayer(logger *zap.Logger, tagOrDigest string, creds []byte) (*AppYamls, error) {
	ref, err := name.ParseReference(tagOrDigest)
	if err != nil {
		return nil, err
	}

	auth := google.NewJSONKeyAuthenticator(string(creds))

	img, err := remote.Image(ref, remote.WithAuth(auth))
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

	tags, err := getTwoTags(logger, ref.Context(), auth, img)
	if err != nil {
		return nil, err
	}

	yamls.Tags = tags

	return yamls, nil
}

func getTwoTags(logger *zap.Logger, repository name.Repository, auth authn.Authenticator, img v1.Image) ([]string, error) {
	var tags []string
	for i := 1; i <= 3; i++ {
		repoImages, err := google.List(repository, google.WithAuth(auth))
		if err != nil {
			return nil, err
		}
		if repoImages == nil {
			return nil, fmt.Errorf("no repo images found")
		}

		digest, err := img.Digest()
		if err != nil {
			return nil, err
		}

		tags = repoImages.Manifests[digest.String()].Tags
		if len(tags) == 2 {
			logger.Debug("get two tags attempt " + fmt.Sprintf("%d", i) + " nailed it")
			return tags, nil
		}
		logger.Debug("get two tags attempt " + fmt.Sprintf("%d", i) + " had " + strings.Join(tags, ", ") + " <- yeah those")
		time.Sleep(5 * time.Second)
	}
	return tags, nil
}

func getTuberYamls(layers []v1.Layer) (*AppYamls, error) {
	for i := len(layers) - 1; i >= 0; i-- {
		size, err := layers[i].Size()
		if err != nil {
			return nil, err
		}
		if size > maxSize {
			continue
		}

		yamls, err := findTuberYamls(layers[i])
		if err != nil {
			return nil, err
		}
		if yamls != nil {
			return yamls, nil
		}
	}

	return nil, fmt.Errorf("tuber yamls not found")
}

func findTuberYamls(layer v1.Layer) (*AppYamls, error) {
	var yamls *AppYamls
	uncompressed, err := layer.Uncompressed()
	if err != nil {
		return nil, err
	}
	archive := tar.NewReader(uncompressed)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return yamls, nil
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

			if yamls == nil {
				yamls = &AppYamls{}
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

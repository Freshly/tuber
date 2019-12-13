package yaml_retrieval

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"tuber/pkg/util"
)

type notTuberLayerError struct {
	message string
}

type layer struct {
	Digest string `json:"digest"`
	Size   int32  `json:"size"`
}

func (e *notTuberLayerError) Error() string { return e.message }

func (l layer) download(image authorizedImage) (yamls []util.Yaml, err error) {
	requestURL := fmt.Sprintf(
		"%s/v2/%s/blobs/%s",
		os.Getenv("REGISTRY_BASE"),
		image.name,
		l.Digest,
	)

	client := &http.Client{}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", image.token))
	res, err := client.Do(req)
	if err != nil {
		return
	}


	gzipped, err := gzip.NewReader(res.Body)
	if err != nil {
		return
	}

	archive := tar.NewReader(gzipped)

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

		yamls = append(yamls, util.Yaml { Filename: header.Name, Content: string(bytes)} )
	}

	return
}


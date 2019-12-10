package yaml_retrieval

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"net/http"
	"os"
)

type registryToken struct {
	Token string `json:"token"`
}

func (r Retriever) getToken() (*registryToken, error) {
	requestURL := fmt.Sprintf(
		"%s/v2/token?scope=repository:%s:pull",
		os.Getenv("AUTH_BASE"),
		r.Image.Name,
	)

	client := &http.Client{}

	req, err := http.NewRequest("GET", requestURL, nil)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("_token", os.Getenv("GCLOUD_TOKEN"))
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	tokenResponse := new(registryToken)
	err = json.Unmarshal(body, &tokenResponse)

	if err != nil {
		return nil, err
	}

	if tokenResponse.Token == "" {
		return nil, fmt.Errorf("no token")
	}

	spew.Dump(tokenResponse)
	return tokenResponse, nil
}
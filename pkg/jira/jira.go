package jira

import (
  "bytes"
  "io/ioutil"
  "net/http"
  "encoding/json"
  "time"
  "math/rand"
  "strconv"
  "regexp"
)

type AuthTokenResponse struct {
  AccessToken string `json:"access_token"`
}

type Deployments struct {
  Deployments []jiraDeploymentBody `json:"deployments"`
}

type jiraDeploymentBody struct {
  SchemaVersion string `json:"schemaVersion"`
  DeploymentSequenceNumber string `json:"deploymentSequenceNumber"`
  UpdateSequenceNumber string `json:"updateSequenceNumber"`
  IssueKeys []string `json:"issueKeys"`
  DisplayName string `json:"displayName"`
  Url string `json:"url"`
  Description string `json:"description"`
  LastUpdated string `json:"lastUpdated"`
  Label string `json:"label"`
  State string `json:"state"`
  Pipeline jiraDeploymentPipeline `json:"pipeline"`
  Environment jiraDeploymentEnvironment `json:"environment"`

}

type jiraDeploymentPipeline struct {
  Id string `json:"id"`
  DisplayName string `json:"displayName"`
  Url string `json:"url"`
}

type jiraDeploymentEnvironment struct {
  Id string `json:"id"`
  DisplayName string `json:"displayName"`
  Type string `json:"type"`
}

type GithubCommits struct {
	Commits []CommitInfo `json:"commits"`
 }

type CommitInfo struct {
 	Commit Commit `json:"commit"`
}

type Commit struct {
	Message string `json:"message"`
}

func GetAuthToken() string {
  var tokenBodyData = []byte(`{
    "audience": "api.atlassian.com",
    "grant_type": "client_credentials",
    "client_id": "ZzSLEhbVeQWkQC8P18YHndHrgBirs19W",
    "client_secret": "mewXlfjQHi3mvDhA57UqBOnBkwXmwxrBBBi5rsSsvfH6yAK-SWTVVK7FBt8PFMv_"
  }`)

  request, error := http.NewRequest("POST", "https://api.atlassian.com/oauth/token", bytes.NewBuffer(tokenBodyData))
  request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  client := &http.Client{}
  response, error := client.Do(request)

  if error != nil { panic(error) }
  defer response.Body.Close()

  var parsedResponse AuthTokenResponse
  body, _ := ioutil.ReadAll(response.Body)
  json.Unmarshal([]byte(body), &parsedResponse)

  return parsedResponse.AccessToken
}

func GetGithubCommitMessages(githubRepo string, oldSha string, newSha string) []string {
	githubAuthToken := "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"
	requestUrl := "https://api.github.com/repos" + githubRepo + "/compare/" + oldSha + "..." + newSha

	request, error := http.NewRequest("GET", requestUrl, nil)
  request.Header.Set("Authorization", "token " + githubAuthToken)
  client := &http.Client{}
  response, error := client.Do(request)

  if error != nil { panic(error) }
  defer response.Body.Close()

  var parsedResponse GithubCommits
  body, _ := ioutil.ReadAll(response.Body)

  json.Unmarshal([]byte(body), &parsedResponse)

	output := make([]string, len(parsedResponse.Commits))
	r, _ := regexp.Compile("[A-Z]+\\-\\d+")

	for _, a := range parsedResponse.Commits {
		storyName := r.FindString(a.Commit.Message)

		if storyName != "" {
			output = append(output, storyName)
		}
	}

  return output
}

func PushJiraDeployment(deploymentUrl string, IssueKeys []string) {
  var cloudId = "a647dcbc-2075-4f2b-bb98-98995953e33f"

  pipeline :=  jiraDeploymentPipeline{
    Id: "Freshly/create-review-app",
    DisplayName: "Tuber Pipeline",
    Url: "https://api.github.com/repos/Freshly/create-review-app/actions/runs/1",
  }
  environment := jiraDeploymentEnvironment{
    Id: "Test",
    DisplayName: "staging",
    Type: "development",
  }

  rand.Seed(time.Now().UnixNano())
  deploymentSequenceNumber := strconv.FormatInt(int64(rand.Intn(1000)), 10) // will be replaced

  deployment := jiraDeploymentBody{
    SchemaVersion: "1.0",
    DeploymentSequenceNumber: deploymentSequenceNumber,
    UpdateSequenceNumber: deploymentSequenceNumber,
    IssueKeys: []string{"HACK-6"},
    DisplayName: "Github Diff",
    Url: deploymentUrl,
    Description: "Test DESCRIPTION",
    LastUpdated: time.Now().Format(time.RFC3339),
    Label: "TEST LABEL",
    State: "successful",
    Pipeline: pipeline,
    Environment: environment,
  }

  var deploymentsBodyData = Deployments{ Deployments: []jiraDeploymentBody{deployment} }
  body, _ := json.Marshal(deploymentsBodyData)

  request, _ := http.NewRequest("POST", "https://api.atlassian.com/jira/deployments/0.1/cloud/" + cloudId + "/bulk", bytes.NewBuffer(body))
  request.Header.Set("Content-Type", "application/json; charset=UTF-8")
  request.Header.Set("Accept", "application/json; charset=UTF-8")
  request.Header.Set("Authorization", "Bearer " + GetAuthToken())

  client := &http.Client{}
  response, error := client.Do(request)

  if error != nil { panic(error) }
  defer response.Body.Close()
}

// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type AppInput struct {
	Name    string `json:"name"`
	IsIstio bool   `json:"isIstio"`
}

type Resource struct {
	Encoded string `json:"encoded"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
}

type ReviewAppsConfig struct {
	Enabled bool        `json:"enabled"`
	Vars    []*Tuple    `json:"vars"`
	Skips   []*Resource `json:"skips"`
}

type TuberApp struct {
	CloudSourceRepo  string            `json:"cloudSourceRepo"`
	ImageTag         string            `json:"imageTag"`
	Name             string            `json:"name"`
	Paused           bool              `json:"paused"`
	Repo             string            `json:"repo"`
	RepoHost         string            `json:"repoHost"`
	RepoPath         string            `json:"repoPath"`
	ReviewApp        bool              `json:"reviewApp"`
	ReviewAppsConfig *ReviewAppsConfig `json:"reviewAppsConfig"`
	SlackChannel     string            `json:"slackChannel"`
	StateResources   []*Resource       `json:"stateResources"`
	Tag              string            `json:"tag"`
	TriggerID        string            `json:"triggerID"`
	Vars             []*Tuple          `json:"vars"`
}

type Tuple struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

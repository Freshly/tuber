// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type AppInput struct {
	Name     string `json:"name"`
	IsIstio  bool   `json:"isIstio"`
	ImageTag string `json:"imageTag"`
	Paused   *bool  `json:"paused"`
}

type CreateReviewAppInput struct {
	Name       string `json:"name"`
	BranchName string `json:"branchName"`
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

type SetTupleInput struct {
	Name  string `json:"name"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type State struct {
	Current  []*Resource `json:"Current"`
	Previous []*Resource `json:"Previous"`
}

type TuberApp struct {
	CloudSourceRepo  string            `json:"cloudSourceRepo"`
	ImageTag         string            `json:"imageTag"`
	Name             string            `json:"name"`
	Paused           bool              `json:"paused"`
	ReviewApp        bool              `json:"reviewApp"`
	ReviewAppsConfig *ReviewAppsConfig `json:"reviewAppsConfig"`
	SlackChannel     string            `json:"slackChannel"`
	SourceAppName    string            `json:"sourceAppName"`
	State            *State            `json:"state"`
	TriggerID        string            `json:"triggerID"`
	Vars             []*Tuple          `json:"vars"`
	ReviewApps       []*TuberApp       `json:"reviewApps"`
	Env              []*Tuple          `json:"env"`
}

type Tuple struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

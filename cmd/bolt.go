package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/spf13/cobra"
)

var bolterCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "bolter",
	Short:        "boltss",
	RunE:         bolter,
}

func bolter(cmd *cobra.Command, args []string) error {
	db, err := db()
	defer db.Close()

	configApps, err := getallconfigapps()
	if err != nil {
		return err
	}

	repos, err := k8s.GetConfigResource("tuber-repos", "tuber", "configmap")
	if err != nil {
		return err
	}

	pauses, err := k8s.GetConfigResource("tuber-app-pauses", "tuber", "configmap")
	if err != nil {
		return err
	}

	var reviewAppTriggers *k8s.ConfigResource

	reviewAppsEnabled := true

	if reviewAppsEnabled {
		reviewAppTriggers, err = k8s.GetConfigResource("tuber-review-triggers", "tuber", "configmap")
		if err != nil {
			return err
		}
	}

	var totalDBDuration int64

	for _, configApp := range configApps {
		appState, err := currentState(configApp)

		starttime := time.Now()
		app := &model.TuberApp{
			Tag:          configApp.Tag,
			ImageTag:     configApp.ImageTag,
			Repo:         configApp.Repo,
			RepoPath:     configApp.RepoPath,
			RepoHost:     configApp.RepoHost,
			Name:         configApp.Name,
			ReviewApp:    configApp.ReviewApp,
			SlackChannel: "",
			Vars:         []*model.Tuple{},
		}

		cloudrepo := cloudrepo(configApp, repos.Data)
		var triggerid string
		rac := model.ReviewAppsConfig{}
		if err != nil {
			return err
		}
		var sourceAppName string
		if configApp.ReviewApp {
			triggerid = reviewAppTriggers.Data[configApp.Name]
			sourceAppName = "qa-replicated"
		} else {
			rac.Enabled = true
			rac.Vars = []*model.Tuple{}
			rac.Skips = []*model.Resource{}
		}
		var paused bool
		if pauses.Data[app.Name] != "" {
			parseboold, err := strconv.ParseBool(pauses.Data[configApp.Name])
			if err != nil {
				return err
			}
			paused = parseboold
		}
		app.Paused = paused
		app.CloudSourceRepo = cloudrepo
		app.TriggerID = triggerid
		app.State = appState
		app.ReviewAppsConfig = &rac
		app.SourceAppName = sourceAppName

		err = db.Save(app)
		if err != nil {
			return err
		}
		dur := time.Since(starttime)
		totalDBDuration += int64(dur)
	}

	okletstryit, err := db.App("qa-replicated-pig-399-close-isnewexpressdesign-experiment")
	if err != nil {
		return err
	}
	fmt.Println(okletstryit)
	fmt.Println("time excluding kubectl data gathering: " + time.Duration(totalDBDuration).String())

	fmt.Println("now for filtering")
	matchingApps, err := db.AppsForTag("gcr.io/freshly-docker/tuber:start-bolting")
	for _, a := range matchingApps {
		fmt.Println(a.Name)
	}

	return nil
}

func cloudrepo(a *model.TuberApp, data map[string]string) string {
	for k, v := range data {
		if v == a.ImageTag {
			return k
		}
	}
	return data[a.ImageTag]
}

func init() {
	rootCmd.AddCommand(bolterCmd)
}

//
//
//
//
//
// mostly copied from releaser cus this is the most temporary nonsense ever
type managedResources []managedResource

type rawState struct {
	Resources     managedResources `json:"resources"`
	PreviousState managedResources `json:"previousState"`
}

type managedResource struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Encoded string `json:"encoded"`
}

func currentState(app *model.TuberApp) (*model.State, error) {
	stateName := "tuber-state-" + app.Name
	exists, err := k8s.Exists("configMap", stateName, app.Name)

	if err != nil {
		return nil, err
	}

	if !exists {
		createErr := k8s.Create(app.Name, "configmap", stateName, `--from-literal=state=`)
		if createErr != nil {
			return nil, err
		}
	}

	stateResource, err := k8s.GetConfigResource(stateName, app.Name, "ConfigMap")
	if err != nil {
		return nil, err
	}

	rawStateData := stateResource.Data["state"]

	var stateData rawState
	if rawStateData != "" {
		unmarshalErr := json.Unmarshal([]byte(rawStateData), &stateData)
		if unmarshalErr != nil {
			return nil, err
		}
	}

	var current []*model.Resource
	var previous []*model.Resource

	for _, prev := range stateData.PreviousState {
		previous = append(previous, &model.Resource{
			Encoded: prev.Encoded,
			Kind:    prev.Kind,
			Name:    prev.Name,
		})
	}

	for _, curr := range stateData.Resources {
		current = append(current, &model.Resource{
			Encoded: curr.Encoded,
			Kind:    curr.Kind,
			Name:    curr.Name,
		})
	}

	return &model.State{
		Current:  current,
		Previous: previous,
	}, nil
}

//
//
// ^ mostly copied from releaser cus this is the most temporary nonsense ever

func getallconfigapps() ([]*model.TuberApp, error) {
	reviewAppsConfig, err := k8s.GetConfigResource("tuber-review-apps", "tuber", "ConfigMap")
	if err != nil {
		return nil, err
	}
	sourceAppsConfig, err := k8s.GetConfigResource("tuber-apps", "tuber", "ConfigMap")
	if err != nil {
		return nil, err
	}
	reviewApps, err := toTuberApps(reviewAppsConfig.Data, true)
	if err != nil {
		return nil, err
	}
	sourceApps, err := toTuberApps(sourceAppsConfig.Data, false)
	if err != nil {
		return nil, err
	}
	var configapps []*model.TuberApp
	for _, app := range reviewApps {
		configapps = append(configapps, app)
	}

	for _, app := range sourceApps {
		configapps = append(configapps, app)
	}
	return configapps, nil
}

func toTuberApps(data map[string]string, reviewApps bool) ([]*model.TuberApp, error) {
	var apps []*model.TuberApp
	for name, imageTag := range data {
		split := strings.SplitN(imageTag, ":", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("error parsing tuber app %s", name)
		}

		repoSplit := strings.SplitN(split[0], "/", 2)
		if len(repoSplit) != 2 {
			return nil, fmt.Errorf("error parsing tuber app %s", name)
		}
		apps = append(apps, &model.TuberApp{
			Name:      name,
			ImageTag:  imageTag,
			Tag:       split[1],
			Repo:      split[0],
			RepoPath:  repoSplit[1],
			RepoHost:  repoSplit[0],
			ReviewApp: reviewApps,
		})
	}
	return apps, nil
}

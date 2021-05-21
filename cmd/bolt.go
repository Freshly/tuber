package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/freshly/tuber/graph/model"
	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bolterCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "bolter",
	Short:        "pulls remote database to local (CURRENTLY FROM CONFIGMAPS)",
	RunE:         bolter,
}

func bolter(cmd *cobra.Command, args []string) error {
	db, err := db()
	if err != nil {
		return err
	}
	defer db.Close()

	return pullLocalDB(db)
}

func pullLocalDB(db *core.DB) error {
	fmt.Println("pulling db from configmaps, takes a sec")
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

	for _, configApp := range configApps {
		appState, err := currentState(configApp)

		app := &model.TuberApp{
			ImageTag:     configApp.ImageTag,
			Name:         configApp.Name,
			ReviewApp:    configApp.ReviewApp,
			SlackChannel: "",
			Vars:         []*model.Tuple{},
		}

		cloudrepo, err := cloudrepo(app, repos.Data)
		if err != nil {
			return err
		}
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
			rac.Enabled = reviewAppsEnabled
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

		err = db.SaveApp(app)
		if err != nil {
			return err
		}
	}

	fmt.Println("done pulling db")
	return nil
}

func cloudrepo(a *model.TuberApp, data map[string]string) (string, error) {
	sourceAppTagGCRRef, err := name.ParseReference(a.ImageTag)
	if err != nil {
		return "", err
	}

	repo := strings.Split(sourceAppTagGCRRef.Name(), ":")[0]

	for k, v := range data {
		if v == repo {
			return k, nil
		}
	}
	return "", nil
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
	sourceAppsConfig, err := k8s.GetConfigResource("tuber-apps", "tuber", "ConfigMap")
	if err != nil {
		return nil, err
	}

	sourceApps, err := toTuberApps(sourceAppsConfig.Data, false)
	if err != nil {
		return nil, err
	}
	var configapps []*model.TuberApp

	for _, app := range sourceApps {
		configapps = append(configapps, app)
	}

	if viper.GetBool("reviewapps-enabled") {
		reviewAppsConfig, err := k8s.GetConfigResource("tuber-review-apps", "tuber", "ConfigMap")
		if err != nil {
			return nil, err
		}

		reviewApps, err := toTuberApps(reviewAppsConfig.Data, true)
		if err != nil {
			return nil, err
		}

		for _, app := range reviewApps {
			configapps = append(configapps, app)
		}
	}

	return configapps, nil
}

func toTuberApps(data map[string]string, reviewApps bool) ([]*model.TuberApp, error) {
	var apps []*model.TuberApp
	for name, imageTag := range data {
		apps = append(apps, &model.TuberApp{
			Name:      name,
			ImageTag:  imageTag,
			ReviewApp: reviewApps,
		})
	}
	return apps, nil
}

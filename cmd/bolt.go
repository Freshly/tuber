package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/freshly/tuber/pkg/core"
	"github.com/freshly/tuber/pkg/k8s"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

var bolterCmd = &cobra.Command{
	SilenceUsage: true,
	Use:          "bolter",
	Short:        "boltss",
	RunE:         bolter,
}

func bolter(cmd *cobra.Command, args []string) error {
	var path string
	if _, err := os.Stat("/etc/tuber-bolt"); os.IsNotExist(err) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = wd + "/localbolt"
		_ = os.Remove(path)
	} else {
		path = "/etc/tuber-bolt/db"
	}
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	// It's a common pattern to call this function for all your top-level buckets after you open your database
	// so you can guarantee that they exist for future transactions. shrug emoji
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("apps"))
		if err != nil {
			return err
		}
		return nil
	})

	data := core.NewData(db)

	configApps, err := core.SourceAndReviewApps()
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
		app := &core.TuberApp{
			Tag:          configApp.Tag,
			ImageTag:     configApp.ImageTag,
			Repo:         configApp.Repo,
			RepoPath:     configApp.RepoPath,
			RepoHost:     configApp.RepoHost,
			Name:         configApp.Name,
			ReviewApp:    configApp.ReviewApp,
			SlackChannel: "",
			Vars:         make(map[string]string),
		}

		cloudrepo := cloudrepo(configApp, repos.Data)
		var triggerid string
		rac := core.ReviewAppsConfig{}
		if err != nil {
			return err
		}
		var sourceAppName string
		if configApp.ReviewApp {
			triggerid = reviewAppTriggers.Data[configApp.Name]
			sourceAppName = "qa-replicated"
		} else {
			rac.Enabled = true
			rac.Vars = make(map[string]string)
			rac.Skips = []core.Resource{}
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

		err = data.CreateApp(app)
		if err != nil {
			return err
		}
		dur := time.Since(starttime)
		totalDBDuration += int64(dur)
	}

	okletstryit, err := data.App("qa-replicated-pig-399-close-isnewexpressdesign-experiment")
	if err != nil {
		return err
	}
	fmt.Println(okletstryit)
	fmt.Println("time excluding kubectl data gathering: " + time.Duration(totalDBDuration).String())

	fmt.Println("now for filtering")
	matchingApps, err := data.AppsForTag("gcr.io/freshly-docker/tuber:start-bolting")
	for _, a := range matchingApps {
		fmt.Println(a.Name)
	}

	return nil
}

func cloudrepo(a core.TuberApp, data map[string]string) string {
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

func currentState(app core.TuberApp) (*core.State, error) {
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

	var current []core.Resource
	var previous []core.Resource

	for _, prev := range stateData.PreviousState {
		previous = append(previous, core.Resource{
			Encoded: prev.Encoded,
			Kind:    prev.Kind,
			Name:    prev.Name,
		})
	}

	for _, curr := range stateData.Resources {
		current = append(current, core.Resource{
			Encoded: curr.Encoded,
			Kind:    curr.Kind,
			Name:    curr.Name,
		})
	}

	return &core.State{
		Current:  current,
		Previous: previous,
	}, nil
}

//
//
// ^ mostly copied from releaser cus this is the most temporary nonsense ever

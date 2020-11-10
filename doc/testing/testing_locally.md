# Testing Tuber Locally

1. Create test tuber app (potatoes w/ test branch)
2. Modify `.env` file to point local tuber at test pubsub topic
3. Run tuber locally with a debugger
4. Trigger Pubsub events on test subscription
5. Profit!


## Point local tuber at test subscription & topic
Add the following lines to the `.env` file

```
# .env
TUBER_PUBSUB_SUBSCRIPTION_NAME=tuber-test
TUBER_PUBSUB_PROJECT=freshly-docker
```


## Install Test Tuber App
If the thing you're testing requires deploying apps or modifying configuration, you should install a test version of the [potatoes tuber app](https://github.com/Freshly/potatoes) with a different name.

First, create a branch off potatoes master
```bash
cd path/to/potatoes
git checkout -b super-test
```

Install test version of potatoes app, using your branch as the final argument:
```bash
tuber apps install potatoes-super-test us.gcr.io/freshly-docker/potatoes super-test
```

## Trigger Pubsub Events
There are several ways to trigger events that will cause tuber to re-build your app. The potatoes cloud build trigger is configured to build on pushes to any branch, so it's ideal for testing.


#### Push a commit on the test branch you created for your test tuber app
```bash
git commit -m "trigger build" --allow-empty
```
You can also edit the HTML in `main.go` and make a real commit.


#### Manually trigger the Pubsub message
Go to the [Subscription Details page](https://console.cloud.google.com/cloudpubsub/subscription/detail/tuber-test?project=freshly-docker) for the test subscription, click "View Messages". In the message viewer tab, repeatedly click "Pull" until you see messages populate. Once you find the message you're looking for,


#### Re-build the cloud build from the Cloud Builds History page.
Go to the [build history](https://console.cloud.google.com/cloud-build/builds?project=freshly-docker) page. At the top of the page, click the "Rebuild" button. When the build finishes, it will trigger another pubsub event.


## Debug Locally
Apologies up front, this only applies to those of us who use VS Code. I believe the steps are similar for the JetBrains Go IDE. If you use Vim, god bless and god speed.

Add a debug configuration for the start command:
```json
{
  "name": "start",
  "type": "go",
  "request": "launch",
  "mode": "debug",
  "program": "${workspaceRoot}/main.go",
  "envFile": "${workspaceFolder}/.env",
  "args": ["start", "-y"]
}
```

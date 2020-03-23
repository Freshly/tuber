# tuber

![logo](logo.png)

## Install
Download the binary file from the latest release: https://github.com/Freshly/tuber/releases/

## What is Tuber?

Tuber is primarily a command line tool used to provide an ergonomic interface to manage a cluster running in Google Kubernetes Engine. It provides commands for common operations such as running commands and getting/setting environment variables.

Tuber is also a continuous delivery service that automates the process of applying newly built container images to applications deployed in a cluster.

## Commands

## Managing Application Environment

### `tuber env set`
Set an environment variable

```bash
USAGE
  $ tuber env set -a <app_name> <key> <value>

OPTIONS
  -a, --app-name the name of the app (required)
```

### `tuber env unset`
Remove an environment variable

```bash
USAGE
  $ tuber env unset -a <app_name> <key> <value>

OPTIONS
  -a, --app-name the name of the app (required)
```

### `tuber env get`

List all environment variables

```bash
USAGE
  $ tuber env get -a <app_name>

OPTIONS
  -a, --app-name the name of the app (required)
```

### `tuber env file`

Bulk apply environment variables based on the contents of a JSON file

```bash
USAGE
  $ tuber env file <path/to/file.json> -a <app_name>

OPTIONS
  -a, --app-name the name of the app (required)
```

## Executing Commands on an Application

### `tuber exec`

Execute a command on your application

```bash
USAGE
  $ tuber exec -a <app_name> <command>

OPTIONS
  -a, --app-name   the name of the app (required)
  -w, --workload   specify a deployment name if it does not match your app name
  -c, --container  specify a container (selects by the deployment name by default)
```


## Kubectl version
`kubectl` version >= 1.15 is required. Check your client version anywhere with `kubectl version`.

## Environment Variables
* `TUBER_PUBSUB_SUBSCRIPTION_NAME`

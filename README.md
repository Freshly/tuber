# tuber

![logo](logo.png)

## Install
### Homebrew
```
brew tap freshly/taps
brew install tuber
```

### Download binary
Download the binary file from the latest release: https://github.com/Freshly/tuber/releases/

## What is Tuber?

Tuber is primarily a command line tool used to provide an ergonomic interface to manage a cluster running in Google Kubernetes Engine. It provides commands for common operations such as running commands and getting/setting environment variables.

Tuber is also a continuous delivery service that automates the process of applying newly built container images to applications deployed in a cluster.

## Usage
See [the documentation](doc/tuber.md) for usage instructions

## Kubectl version
`kubectl` version >= 1.15 is required. Check your client version anywhere with `kubectl version`.

## Environment Variables
* `TUBER_PUBSUB_SUBSCRIPTION_NAME`

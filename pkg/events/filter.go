package events

import (
	"fmt"
	"regexp"
	"tuber/pkg/util"
)

type pendingRelease struct {
	name string
	branch string
}

func filter(e *util.RegistryEvent) (event *pendingRelease) {
	filterRegex := regexp.MustCompile(`us\.gcr\.io\/(.*):(.*)`)
	slicedTag := filterRegex.FindStringSubmatch(e.Tag)
	name := slicedTag[1]
	branch := slicedTag[2]

	if name == "tuber" && branch == "master" {
		return &pendingRelease { name: name, branch: branch }
	} else {
		fmt.Println("Ignoring", name, branch)
		e.Message.Ack()
	}
	return
}
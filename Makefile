VERSION := $(shell git describe --tags --abbrev=0)

GOLDFLAGS += -X main.version=$(VERSION)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

build:
	go build -o tuber $(GOFLAGS) .


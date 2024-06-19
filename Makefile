EXECUTABLE=ambiplayer
VERSION=$(shell git describe --tags --always --long --dirty)
##VERSION = "2.1"
WINDOWS=$(EXECUTABLE)_$(VERSION)_windows_amd64.exe

LDFLAGS = "-s -w -X main.AppVersion=$(VERSION) -H=windowsgui"

.PHONY: all test clean

all: test build ## Build and run tests

test: ## Run unit tests
	go test ./...

build: windows ## Build binaries
	@echo version: $(VERSION)

windows: $(WINDOWS) ## Build for Windows



$(WINDOWS):
	 go build -v -o $(WINDOWS) -ldflags=$(LDFLAGS)  .



clean: ## Remove previous build
	del $(WINDOWS)

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

SHELL := /bin/bash

GOCMD=go
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install

BINARY_NAME=cb-restapigw
BINARY_LINUX=$(BINARY_NAME)
TAGS=$(BINARY_NAME)
CONF=./conf/$(BINARY_NAME).yaml

export CBLOG_ROOT=$(CURDIR)
export CBSTORE_ROOT=$(CURDIR)

.PHONY: all clean build
all: clean test build

.PHONY: clean
clean:
	@echo "==> Cleaning project"
	rm -f $(BINARY_NAME)

.PHONY: build-run run check
build-run: build-dev run
build-check: build-dev check

check:
	@echo "==> Check the configuration file"
	./$(BINARY_NAME) --config $(CONF) check

run:
	@echo "==> Build project"
	./$(BINARY_NAME) --config $(CONF) run

.PHONY: build build-dev
build: static
	@echo "==> Build project"
	$(GOBUILD) -tags "$(TAGS)" -o $(BINARY_NAME) -v $(BUILD_PATH)

build-dev:
	@echo "==> Build project (Dev)"
	$(GOBUILD) -tags "$(TAGS)" -o $(BINARY_NAME) -v $(BUILD_PATH)

.PHONY: build-linux
build-linux:
	@echo "==> Build (Linux-64) project"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -tags "$(TAGS)" -o $(BINARY_LINUX) -v $(BUILD_PATH)

.PHONY: install
install:
	@echo "==> Install project"
	$(GOINSTALL) -tags "$(TAGS)"

.PHONY: static clean-client client 
static: client
	esc -o ${PWD}/pkg/static/static.go -pkg static web/dist

clean-client:
	rm -rf ${PWD}/web/.nuxt ${PWD}/web/dist

client: clean-client
	cd ${PWD}/web; yarn generate
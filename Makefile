MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT := $(dir $(MAKEFILE_PATH))

BINARY_NAME=arc-cleaner
VERSION=""
COMMIT := $(shell git rev-parse --verify HEAD)
ifeq ($(WITH_OS_ARCH_SUFFIX), true)
	OS := $(shell go env GOOS)
	ARCH := $(shell go env GOARCH)
	BINARY_NAME := $(BINARY_NAME)-$(OS)-$(ARCH)
endif

PKG := ./...

DEBUG := true

LD_FLAGS =  -extldflags \"-static\" \
	-X \"github.com/wielewoutlabs/arc-cleaner/cmd.version=$(VERSION)\" \
	-X \"github.com/wielewoutlabs/arc-cleaner/cmd.commit=$(COMMIT)\"

ifeq ($(DEBUG),true)
    BUILD_ARGS = -ldflags "$(LD_FLAGS)"
else
	BUILD_ARGS = -ldflags "-s -w $(LD_FLAGS)"
endif

DOCKER_SOCKET := /var/run/docker.sock
DOCKER_CONFIG := ~/.docker

LOCAL_DEVCONTAINER := false
ifeq ($(LOCAL_DEVCONTAINER),true)
		DEVCONTAINER_TAG := sha-$(shell git rev-parse --short HEAD)
		DEVCONTAINER ?= ghcr.io/wielewoutlabs/arc-cleaner-dev:$(DEVCONTAINER_TAG)
else
		# renovate:
		DEVCONTAINER ?= ghcr.io/wielewoutlabs/arc-cleaner-dev:edge@sha256:378163c77cedeebdaffe188cd4625bc9a4630b68c43c2c054a67a7252e282f66
endif
DEVCONTAINER_NAME := arc-cleaner-dev
DEVCONTAINER_WORKDIR := /go/src/github.com/wielewoutlabs/arc-cleaner/
DEVCONTAINER_RUN := docker run -d --volume $(ROOT):$(DEVCONTAINER_WORKDIR) --volume $(DOCKER_SOCKET):/var/run/docker.sock --volume $(DOCKER_CONFIG):/root/.docker --name $(DEVCONTAINER_NAME) $(DEVCONTAINER)

DEVCONTAINER_EXEC := docker exec -it $(DEVCONTAINER_NAME)

ifeq ($(ROOT),$(DEVCONTAINER_WORKDIR))
	IN_DEVCONTAINER := true
endif

ifeq ($(IN_DEVCONTAINER),true)
	GOPATH ?= /go
else
	GOPATH ?= $(HOME)/go
endif
GOBIN ?= $(GOPATH)/bin
PATH := $(GOBIN):$(PATH)

.PHONY: all
all: help

.PHONY: help
help: ## Show help menu
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} \
		/^[a-zA-Z_-]+:.*?## / { \
			printf "    %-24s%s\n", $$1, $$2 \
		}' $(MAKEFILE_LIST)
ifneq ($(IN_DEVCONTAINER),true)
	@awk 'BEGIN {FS = ":.*?##!dev "} \
		/^[a-zA-Z_-]+:.*?##!dev / { \
			printf "    %-24s%s\n", $$1, $$2 \
		}' $(MAKEFILE_LIST)
else
		@echo ''
		@echo 'Running in devcontainer'
endif
	@echo ''

.PHONY: build
build: ## Build application
	@mkdir -p $(ROOT)/bin
	@go build $(BUILD_ARGS) -o $(ROOT)/bin/$(BINARY_NAME) main.go
ifneq ($(DEBUG),true)
		@upx --best --lzma $(ROOT)/bin/$(BINARY_NAME)
endif

.PHONY: run
run: build ## Run application
	@$(ROOT)/bin/$(BINARY_NAME)

.PHONY: test
test: ## Run all tests
	@go test -v $(PKG)

.PHONY: test-unit
test-unit: ## Run all unit tests
	@go test -v -short $(PKG)

.PHONY: test-acceptance
test-acceptance: ## Run all acceptance tests
	@go test -v ./test/acceptance/...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@go test -covermode=count -coverprofile=profile.cov $(PKG) || true
	@go tool cover -func profile.cov

.PHONY: test-coverage-report
test-coverage-report: test-coverage ## Run tests with coverage and open report
	@go tool cover --html profile.cov

.PHONY: lint
lint: ## Lint code
	@golangci-lint run --config .golangci-lint.yaml $(PKG)

.PHONY: clean
clean: ## Clean up artifacts
	@rm -rf $(ROOT)/bin $(ROOT)/*.cov


ifneq ($(IN_DEVCONTAINER),true)

.PHONY: dev-build
dev-build: ##!dev Build devcontainer
		@docker image build --tag "$(DEVCONTAINER)" -f build/devcontainer/Containerfile .

.PHONY: dev-start
ifeq ($(LOCAL_DEVCONTAINER),true)
dev-start: dev-build
else
dev-start: ##!dev Start devcontainer if it is not running yet
endif
		@if test "$(shell docker ps -aq -f status=running -f name=$(DEVCONTAINER_NAME))" ; then \
			echo "Devcontainer is already running" ; \
		else \
			$(DEVCONTAINER_RUN) tail -f /dev/null ; \
		fi

.PHONY: dev-shell
dev-shell: dev-start ##!dev Jump into a shell within the started devcontainer
		@$(DEVCONTAINER_EXEC) sh

.PHONY: dev-stop
dev-stop: ##!dev Stop devcontainer if it is running
		@if test "$(shell docker ps -aq -f status=running -f name=$(DEVCONTAINER_NAME))" ; then \
			docker kill $(DEVCONTAINER_NAME) ; \
		fi
		@if test "$(shell docker ps -aq -f name=$(DEVCONTAINER_NAME))" ; then \
			docker rm $(DEVCONTAINER_NAME) ; \
		fi

.PHONY: dev-restart
dev-restart: dev-stop dev-start ##!dev Restart devcontainer

endif

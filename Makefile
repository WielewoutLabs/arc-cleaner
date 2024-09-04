MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT := $(dir $(MAKEFILE_PATH))

BINARY_NAME=arc-cleaner
VERSION=""
COMMIT := $(shell git rev-parse --verify HEAD)

PKG := ./...

LD_FLAGS = -ldflags " \
	-s \
	-w \
	-extldflags \"-static\" \
	-X \"github.com/wielewout/arc-cleaner/cmd.version=$(VERSION)\" \
	-X \"github.com/wielewout/arc-cleaner/cmd.commit=$(COMMIT)\" \
	"

DOCKER_SOCKET := /var/run/docker.sock

LOCAL_DEVCONTAINER := false
ifeq ($(LOCAL_DEVCONTAINER),true)
		DEVCONTAINER_TAG := sha-$(shell git rev-parse --short HEAD)
		DEVCONTAINER ?= ghcr.io/wielewout/arc-cleaner-dev:$(DEVCONTAINER_TAG)
else
		DEVCONTAINER ?= ghcr.io/wielewout/arc-cleaner-dev:edge
endif
DEVCONTAINER_NAME := arc-cleaner-dev
DEVCONTAINER_WORKDIR := /go/src/github.com/wielewout/arc-cleaner/
DEVCONTAINER_RUN := docker run -d --volume $(ROOT):$(DEVCONTAINER_WORKDIR) --volume $(DOCKER_SOCKET):/var/run/docker.sock --name $(DEVCONTAINER_NAME) $(DEVCONTAINER)

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
	@go build $(LD_FLAGS) -o $(ROOT)/bin/$(BINARY_NAME) main.go

.PHONY: run
run: build ## Run application
	@$(ROOT)/bin/$(BINARY_NAME)

.PHONY: test
test: ## Run all tests
	@go test -v $(PKG)

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

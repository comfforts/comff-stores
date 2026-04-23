# Git info
CURRENT_TAG   := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "none")
CURRENT_DESC  := $(shell git describe --tags --always 2>/dev/null || echo "none")
HEAD_SHA      := $(shell git rev-parse HEAD)
SHORT_SHA     := $(shell git rev-parse --short HEAD)
BRANCH        := $(shell git rev-parse --abbrev-ref HEAD)

ifeq ($(CURRENT_TAG),none)
  CURRENT_TAG := v0.0.0
endif

# Image tag for docker builds
IMG_TAG := $(strip $(CURRENT_TAG)-$(SHORT_SHA))
PROD_IMG_TAG := $(strip $(CURRENT_TAG))

# Default platform and architecture based on host machine
HOST_UNAME_S := $(shell uname -s)
HOST_UNAME_M := $(shell uname -m)
ifeq ($(HOST_UNAME_S),Darwin)
  ifeq ($(HOST_UNAME_M),arm64)
    DEFAULT_PLATFORM := linux/arm64
    DEFAULT_ARCH := osx-arm64
  else ifeq ($(HOST_UNAME_M),x86_64)
    DEFAULT_PLATFORM := linux/amd64
    DEFAULT_ARCH := osx-x86_64
  else
    DEFAULT_PLATFORM := linux/amd64
    DEFAULT_ARCH := osx-universal2
  endif
else ifeq ($(HOST_UNAME_S),Linux)
  ifeq ($(HOST_UNAME_M),aarch64)
    DEFAULT_PLATFORM := linux/arm64
	DEFAULT_ARCH := linux-aarch64
  else ifeq ($(HOST_UNAME_M),arm64)
    DEFAULT_PLATFORM := linux/arm64
	DEFAULT_ARCH := linux-aarch64
  else ifeq ($(HOST_UNAME_M),x86_64)
    DEFAULT_PLATFORM := linux/amd64
	DEFAULT_ARCH := linux-x64
  else ifeq ($(HOST_UNAME_M),amd64)
    DEFAULT_PLATFORM := linux/amd64
	DEFAULT_ARCH := linux-x64
  else
    DEFAULT_PLATFORM := linux/amd64
	DEFAULT_ARCH := linux-aarch64
  endif
else
  DEFAULT_PLATFORM := linux/amd64
  DEFAULT_ARCH := linux-x64
endif

PLATFORM    ?= $(DEFAULT_PLATFORM)
OS_ARCH    ?= $(DEFAULT_ARCH)
GO_VERSION  ?= 1.25.9

OS_NAME := $(shell go env GOOS)
ARCH_NAME := $(shell go env GOARCH)

# current git repo info
.PHONY: git-info
git-info:
	@echo "CURRENT_TAG = $(CURRENT_TAG)"
	@echo "CURRENT_DESC = $(CURRENT_DESC)"
	@echo "HEAD_SHA = $(HEAD_SHA)"
	@echo "SHORT_SHA = $(SHORT_SHA)"
	@echo "BRANCH = $(BRANCH)"

# current platform info
.PHONY: platform-info
platform-info:
	@echo "HOST_UNAME_S = $(HOST_UNAME_S)"
	@echo "HOST_UNAME_M = $(HOST_UNAME_M)"
	@echo "PLATFORM = $(PLATFORM)"
	@echo "OS_ARCH = $(OS_ARCH)"
	@echo "OS_NAME = $(OS_NAME)"
	@echo "ARCH_NAME = $(ARCH_NAME)"

######## - Proto - #######
# setup proto tools
.PHONY: setup-proto
setup-proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# build service protos
.PHONY: build-proto
build-proto:
	@echo "building latest service protos"
	scripts/build-proto.sh


######## - Network - #######
# setup docker network for local development
.PHONY: network
network:
	@echo "Creating storesnet network if it doesn't exist..."
	@if ! docker network inspect storesnet > /dev/null 2>&1; then \
		docker network create storesnet; \
		echo "Network storesnet created."; \
	else \
		echo "Network storesnet already exists."; \
	fi

######## - Data stores - #######
## Mongo ##
# start mongo - START:bootstrap/secure
.PHONY: start-mongo
start-mongo: network
	@echo "Starting MongoDB..."
	scripts/start-mongo.sh ${START}

# stop mongo
.PHONY: stop-mongo
stop-mongo:
	@echo "Stopping MongoDB cluster..."
	@set -a; . env/mongo3n2.env; set +a; docker-compose -f deploy/stores/mongo/docker-compose-mongo-3-node-2-phase.yml --profile secure down

######## - Services - #######
## Local Stores server ##
.PHONY: start-server
start-server:
	@echo "starting stores server with latest ${HEAD}"
	scripts/start-server.sh stores


.PHONY: run-test
run-test:
	@echo "testing latest ${HEAD}"
	go test --race -v ./...

.PHONY: build-exec
build-exec:
	@echo "building latest executables for ${HEAD}"
	scripts/build-exec.sh
	scripts/build-exec.sh darwin arm64
	scripts/build-exec.sh linux arm64
	scripts/build-exec.sh linux amd64

.PHONY: build-docker
build-docker:
	@echo "building docker image for ${HEAD}"
	scripts/build-docker.sh

.PHONY: run-docker
run-docker:
	@echo "running docker image build for ${HEAD}"
	scripts/run-docker.sh

.PHONY: build-docker-test
build-docker-test:
	@echo "building docker image for ${HEAD}"
	scripts/build-test.sh

.PHONY: run-docker-test
run-docker-test:
	@echo "running docker image build for ${HEAD}"
	scripts/run-test.sh

.PHONY: start-agent
start-agent:
	@echo "starting agent with latest ${HEAD}"
	rm -rf cmd/cli/data/raft
	cd cmd/cli && go run comffstore.go --data-dir data --bootstrap true

.PHONY: build-agent-exec
build-agent-exec:
	@echo "building latest executables for ${HEAD}"
	scripts/build-agent-exec.sh
	scripts/build-agent-exec.sh darwin arm64
	scripts/build-agent-exec.sh linux arm64
	scripts/build-agent-exec.sh linux amd64

.PHONY: build-agent-docker
build-agent-docker:
	@echo "building docker image for ${HEAD}"
	scripts/build-agent-docker.sh

.PHONY: run-agent-docker
run-agent-docker:
	@echo "running docker image build for ${HEAD}"
	scripts/run-agent-docker.sh



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

# Services in compose
STORES_SERVICE        := stores-server
STORES_DEBUG_SERVICE  := stores-server-debug

# Docker compose acronyms
STORES_COMPOSE_FILE  := deploy/stores/docker-compose-stores.yaml
STORES_COMPOSE       := docker compose -f $(STORES_COMPOSE_FILE)

# Profiles
PROD_PROFILE   := prod
DEBUG_PROFILE  := debug

# Build targets from Dockerfile
DOCKER_TARGET_PROD  ?= runtime
DOCKER_TARGET_DEBUG ?= runtime-debug

# Buildkit convenience
BUILD_ENV := DOCKER_BUILDKIT=1

# Image tag for docker builds
IMG_TAG := $(strip $(CURRENT_TAG)-$(SHORT_SHA))
PROD_IMG_TAG := $(strip $(CURRENT_TAG))

# Carriage return character (for stripping CRLF contamination)
CR := $(shell printf '\r')
# linux/arm64 -> arm64 ; linux/amd64 -> amd64
ARCH_SUFFIX := $(strip $(notdir $(PLATFORM)))
# Clean PLATFORM (remove CR, then strip whitespace)
ARCH_SUFFIX_CLEAN := $(strip $(subst $(CR),,$(ARCH_SUFFIX)))

# Image naming (arch-tagged) for multi-arch builds
STORES_IMAGE  := $(STORES_SERVICE)-$(ARCH_SUFFIX_CLEAN):dev-$(IMG_TAG)
STORES_PROD_IMAGE  := $(STORES_SERVICE)-$(ARCH_SUFFIX_CLEAN):prod-$(PROD_IMG_TAG)
STORES_LOCAL_IMAGE  := $(STORES_SERVICE)-$(ARCH_SUFFIX_CLEAN):devLocal-$(IMG_TAG)
STORES_DEBUG_IMAGE  := $(STORES_SERVICE)-$(ARCH_SUFFIX_CLEAN):debug-$(IMG_TAG)
STORES_DEBUG_PROD_IMAGE  := $(STORES_SERVICE)-$(ARCH_SUFFIX_CLEAN):debug-$(PROD_IMG_TAG)
STORES_DEBUG_LOCAL_IMAGE  := $(STORES_SERVICE)-$(ARCH_SUFFIX_CLEAN):debugLocal-$(IMG_TAG)

# Export for compose interpolation
export IMG_TAG GO_VERSION ORT_VERSION PLATFORM
export DOCKER_TARGET_PROD DOCKER_TARGET_DEBUG
export STORES_IMAGE STORES_PROD_IMAGE STORES_LOCAL_IMAGE
export STORES_DEBUG_IMAGE STORES_DEBUG_PROD_IMAGE STORES_DEBUG_LOCAL_IMAGE

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
	@echo "Creating comffnet network if it doesn't exist..."
	@if ! docker network inspect comffnet > /dev/null 2>&1; then \
		docker network create comffnet; \
		echo "Network comffnet created."; \
	else \
		echo "Network comffnet already exists."; \
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

######## - Stores Service: docker - #######
# build stores service debug docker image with docker compose
.PHONY: build-stores-debug
build-stores-debug:
	@echo "building debug mode stores service docker image with IMG_TAG=$(IMG_TAG) on BRANCH=$(BRANCH) PLATFORM=$(PLATFORM)"
	DOCKER_BUILDKIT=1 docker compose -f deploy/stores/docker-compose-stores.yaml --profile debug --progress plain build --no-cache stores-server-debug
# 	$(BUILD_ENV) $(STORES_COMPOSE) --profile $(DEBUG_PROFILE) --progress plain build $(STORES_DEBUG_SERVICE)


# build stores service docker image with docker compose
.PHONY: build-stores
build-stores:
	@echo "building stores service docker image with IMG_TAG=$(IMG_TAG) on BRANCH=$(BRANCH) PLATFORM=$(PLATFORM)"
	DOCKER_BUILDKIT=1 docker compose -f deploy/stores/docker-compose-stores.yaml --profile prod --progress plain build --no-cache stores-server
# 	$(BUILD_ENV) $(STORES_COMPOSE) --profile $(PROD_PROFILE) --progress plain build $(STORES_SERVICE)

# start stores service debug docker image with docker compose
.PHONY: start-stores-debug
start-stores-debug: build-stores-debug network
	@echo "starting debug mode stores service docker image with IMG_TAG=$(IMG_TAG) on BRANCH=$(BRANCH) PLATFORM=$(PLATFORM)"
	$(STORES_COMPOSE) --profile $(DEBUG_PROFILE) up -d $(STORES_DEBUG_SERVICE)

# stop stores service debug docker image with docker compose
.PHONY: stop-stores-debug
stop-stores-debug:
	@echo "stopping debug mode stores service"
	$(STORES_COMPOSE) --profile $(DEBUG_PROFILE) down

# build & start stores service with docker compose
.PHONY: start-stores
start-stores: build-stores network
	@echo "starting stores service IMG_TAG=$(IMG_TAG) BRANCH=$(BRANCH) PLATFORM=$(PLATFORM)"
	$(STORES_COMPOSE) --profile $(PROD_PROFILE) up -d $(STORES_SERVICE)

# stop docker composed stores service
.PHONY: stop-stores
stop-stores:
	@echo "stopping stores service"
	$(STORES_COMPOSE) --profile $(PROD_PROFILE) down

.PHONY: run-test
run-test:
	@echo "testing latest ${HEAD}"
	go test --race -v ./...

######## - Stores Service: K8s, kind - #######
.PHONY: build-stores-creds-k
build-stores-creds-k:
	@echo "building stores creds manifests for k8s cluster"
	kubectl -n comff create configmap stores-policy --from-file=policy.csv=./cmd/servers/stores/policies/policy.csv --dry-run=client -o yaml > k8s/stores/stores-policy.yaml
	kubectl -n comff create configmap stores-config --from-env-file=./env/stores-config.env --dry-run=client -o yaml > k8s/stores/stores-config.yaml
	kubectl -n comff create secret generic stores-secret --from-env-file=./env/stores-secret.env --dry-run=client -o yaml > k8s/stores/stores-secret.yaml

# Load stores service & dependencies images into kind cluster
.PHONY: load-img-k
load-img-k:
	@echo "loading container image into kind k8s cluster"
	kind load docker-image stores-server-arm64:prod-v0.0.2 --name comff

# Load stores service & dependencies images into kind cluster
.PHONY: load-stores-imgs-k
load-stores-imgs-k:
	@echo "loading container images into kind k8s cluster"
	kind load docker-image $(STORES_PROD_IMAGE) --name comff
	kind load docker-image $(STORES_DEBUG_PROD_IMAGE) --name comff

.PHONY: set-stores-k
set-stores-k:
	@echo "setting up stores in k8s cluster"
	kubectl apply -k k8s/stores

.PHONY: rm-stores-k
rm-stores-k:
	@echo "removing stores from k8s cluster"
	kubectl delete -k k8s/stores


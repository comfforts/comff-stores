OS_NAME := $(shell go env GOOS)
ARCH_NAME := $(shell go env GOARCH)
your_env = OS: $(OS_NAME) ARCH: $(ARCH_NAME)

LINUX_OS = "linux"
LINUX_ARCH = "amd64"

HEAD := $(shell git rev-parse --short HEAD)

.PHONY: envars
envars: 
	@echo $(your_env)

.PHONY: build-proto
build-proto:
	@echo "building latest proto for ${HEAD}"
	scripts/build-proto.sh

.PHONY: run-test
run-test:
	@echo "testing latest ${HEAD}"
	cd pkg/utils/geohash && go test --race -v .
	cd pkg/services/store && go test --race -v .
	cd pkg/services/geocode && go test --race -v .
	cd pkg/services/filestorage && go test --race -v .
	cd pkg/jobs && go test --race -v .
	cd pkg/config && go test --race -v .
	cd internal/server && go test --race -v .

.PHONY: start-server
start-server:
	@echo "starting server with latest ${HEAD}"
	cd cmd/store && go run store.go

.PHONY: build-exec
build-exec:
	@echo "building latest executables for ${HEAD}"
	scripts/build-exec.sh
	scripts/build-exec.sh darwin amd64
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


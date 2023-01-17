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
	go test --race -v ./...

.PHONY: start-server
start-server:
	@echo "starting server with latest ${HEAD}"
	cd cmd/store && go run store.go

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



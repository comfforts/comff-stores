#!/bin/bash

docker build --progress=plain --build-arg BUILD_OS=darwin --build-arg BUILD_ARCH=arm64 -t comff-stores:test -f deploy/docker/test/Dockerfile .
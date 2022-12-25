#!/bin/bash

docker build --progress=plain --build-arg BUILD_OS=darwin --build-arg BUILD_ARCH=amd64 -t comff-stores:test -f docker/test/Dockerfile .
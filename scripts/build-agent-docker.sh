#!/bin/bash

CURR_SHA=$(git rev-parse --short HEAD)
CURR_BRANCH=$(git rev-parse --abbrev-ref HEAD)
BASEDIR=$(dirname "$0")
echo $BASEDIR

LOCAL_OS=$(go env GOOS)
LOCAL_ARCH=$(go env GOARCH)
echo "local OS: ${LOCAL_OS} for local arch: ${LOCAL_ARCH}"

BUILD_OS="linux"
BUILD_ARCH=$LOCAL_ARCH

cd ${BASEDIR}/../
echo "building docker image for rev: ${CURR_SHA}, branch: ${CURR_BRANCH}, arch: ${BUILD_OS}/${BUILD_ARCH}"
docker build --progress=plain --build-arg BUILD_OS=${BUILD_OS} --build-arg BUILD_ARCH=${BUILD_ARCH} -t github.com/comfforts/comff-stores:${CURR_SHA} -f deploy/docker/agent/Dockerfile .